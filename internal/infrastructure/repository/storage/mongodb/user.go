package mongodb

import (
	"context"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/dto"
	"github.com/Borislavv/video-streaming/internal/domain/errtype"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	diinterface "github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	"github.com/Borislavv/video-streaming/internal/domain/vo"
	queryinterface "github.com/Borislavv/video-streaming/internal/infrastructure/repository/query/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

const UserCollection = "users"

var (
	UserNotFoundByIdError    = errtype.NewEntityNotFoundError("mongo", "user", "id")
	UserNotFoundByEmailError = errtype.NewEntityNotFoundError("mongo", "user", "email")
	UserInsertingFailedError = errtype.NewInternalValidationError("unable to store 'user' or get inserted 'id'")
	UserWasNotDeletedError   = errtype.NewInternalValidationError("user was not deleted")
)

type UserRepository struct {
	db      *mongo.Collection
	mu      *sync.Mutex
	logger  loggerinterface.Logger
	timeout time.Duration
}

func NewUserRepository(serviceContainer diinterface.ServiceContainer) (*UserRepository, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	mongodb, err := serviceContainer.GetMongoDatabase()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	cfg, err := serviceContainer.GetConfig()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	timeout, err := time.ParseDuration(cfg.MongoTimeout)
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &UserRepository{
		db:      mongodb.Collection(UserCollection),
		logger:  loggerService,
		mu:      &sync.Mutex{},
		timeout: timeout,
	}, nil
}

func (r *UserRepository) FindOneByID(ctx context.Context, q queryinterface.FindOneUserByID) (user *agg.User, err error) {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	filter := bson.M{"_id": q.GetID().Value}

	user = &agg.User{}
	if err = r.db.FindOne(qCtx, filter).Decode(user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, r.logger.InfoPropagate(UserNotFoundByIdError)
		}
		return nil, r.logger.ErrorPropagate(err)
	}

	return user, nil
}

func (r *UserRepository) FindOneByEmail(ctx context.Context, q queryinterface.FindOneUserByEmail) (user *agg.User, err error) {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	filter := bson.M{"email": q.GetEmail()}

	user = &agg.User{}
	if err = r.db.FindOne(qCtx, filter).Decode(user); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, r.logger.InfoPropagate(UserNotFoundByEmailError)
		}
		return nil, r.logger.ErrorPropagate(err)
	}

	return user, nil
}

func (r *UserRepository) Insert(ctx context.Context, user *agg.User) (*agg.User, error) {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.db.InsertOne(qCtx, user, options.InsertOne())
	if err != nil {
		return nil, r.logger.ErrorPropagate(err)
	}

	if oID, ok := res.InsertedID.(primitive.ObjectID); ok {
		q := dto.NewUserGetRequestDTO(vo.NewID(oID), "")
		return r.FindOneByID(qCtx, q)
	}

	return nil, r.logger.CriticalPropagate(UserInsertingFailedError)
}

func (r *UserRepository) Update(ctx context.Context, user *agg.User) (*agg.User, error) {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.db.UpdateByID(qCtx, user.ID.Value, bson.M{"$set": user})
	if err != nil {
		return nil, r.logger.ErrorPropagate(err)
	}

	// check the record is really updated
	if res.ModifiedCount > 0 {
		q := dto.NewUserGetRequestDTO(user.ID, "")
		return r.FindOneByID(qCtx, q)
	}

	// if changes is not exists, then return the original data
	return user, nil
}

func (r *UserRepository) Remove(ctx context.Context, user *agg.User) error {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.db.DeleteOne(qCtx, bson.M{"_id": user.ID.Value})
	if err != nil {
		return r.logger.ErrorPropagate(err)
	}

	if res.DeletedCount == 0 { // checking the user was deleted
		return r.logger.CriticalPropagate(UserWasNotDeletedError)
	}

	return nil
}
