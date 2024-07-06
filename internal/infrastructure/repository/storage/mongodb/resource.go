package mongodb

import (
	"context"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/dto"
	"github.com/Borislavv/video-streaming/internal/domain/errtype"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	"github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	"github.com/Borislavv/video-streaming/internal/domain/vo"
	queryinterface "github.com/Borislavv/video-streaming/internal/infrastructure/repository/query/interface"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sync"
	"time"
)

const ResourcesCollection = "resources"

var (
	ResourceNotFoundByIdError    = errtype.NewEntityNotFoundError("mongo", "resource", "id")
	ResourceInsertingFailedError = errtype.NewInternalRepositoryError("unable to store 'resource' or retrieve inserted 'id'")
)

type ResourceRepository struct {
	db      *mongo.Collection
	mu      *sync.Mutex
	logger  loggerinterface.Logger
	timeout time.Duration
}

func NewResourceRepository(serviceContainer diinterface.ServiceContainer) (*ResourceRepository, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	database, err := serviceContainer.GetMongoDatabase()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	config, err := serviceContainer.GetConfig()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	timeout, err := time.ParseDuration(config.MongoTimeout)
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &ResourceRepository{
		db:      database.Collection(ResourcesCollection),
		logger:  loggerService,
		mu:      &sync.Mutex{},
		timeout: timeout,
	}, nil
}

func (r *ResourceRepository) FindOneByID(ctx context.Context, q queryinterface.FindOneResourceByID) (*agg.Resource, error) {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	filter := bson.M{
		"_id":      q.GetID().Value,
		"user._id": q.GetUserID().Value,
	}

	resourceAgg := &agg.Resource{}
	if err := r.db.FindOne(qCtx, filter).Decode(resourceAgg); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, r.logger.InfoPropagate(ResourceNotFoundByIdError)
		}
		return nil, r.logger.ErrorPropagate(err)
	}

	return resourceAgg, nil
}

func (r *ResourceRepository) Insert(ctx context.Context, resource *agg.Resource) (*agg.Resource, error) {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.db.InsertOne(qCtx, resource, options.InsertOne())
	if err != nil {
		return nil, r.logger.ErrorPropagate(err)
	}

	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		q := dto.NewResourceGetRequestDTO(vo.ID{Value: oid}, resource.UserID)
		return r.FindOneByID(qCtx, q)
	}

	return nil, r.logger.CriticalPropagate(ResourceInsertingFailedError)
}

func (r *ResourceRepository) Remove(ctx context.Context, resource *agg.Resource) error {
	qCtx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	res, err := r.db.DeleteOne(qCtx, bson.M{"_id": resource.ID.Value})
	if err != nil {
		return r.logger.ErrorPropagate(err)
	}

	if res.DeletedCount == 0 { // checking the resource was deleted
		return r.logger.CriticalPropagate(UserWasNotDeletedError)
	}

	return nil
}
