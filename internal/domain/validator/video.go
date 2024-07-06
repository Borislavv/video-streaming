package validator

import (
	"context"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/dto"
	dtointerface "github.com/Borislavv/video-streaming/internal/domain/dto/interface"
	"github.com/Borislavv/video-streaming/internal/domain/errtype"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	repositoryinterface "github.com/Borislavv/video-streaming/internal/domain/repository/interface"
	"github.com/Borislavv/video-streaming/internal/domain/service/accessor/interface"
	diinterface "github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	validatorinterface "github.com/Borislavv/video-streaming/internal/domain/validator/interface"
	"github.com/Borislavv/video-streaming/internal/domain/vo"
)

const (
	idField         = "id"
	userIDField     = "userID"
	nameField       = "name"
	resourceIDField = "resourceID"
)

type VideoValidator struct {
	ctx                context.Context
	logger             loggerinterface.Logger
	resourceValidator  validatorinterface.Resource
	accessService      accessorinterface.Accessor
	videoRepository    repositoryinterface.Video
	resourceRepository repositoryinterface.Resource
}

func NewVideoValidator(serviceContainer diinterface.ServiceContainer) (*VideoValidator, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	ctx, err := serviceContainer.GetCtx()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	resourceValidatorService, err := serviceContainer.GetResourceValidator()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	accessService, err := serviceContainer.GetAccessService()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	videoRepository, err := serviceContainer.GetVideoRepository()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	resourceRepository, err := serviceContainer.GetResourceRepository()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &VideoValidator{
		ctx:                ctx,
		logger:             loggerService,
		resourceValidator:  resourceValidatorService,
		accessService:      accessService,
		videoRepository:    videoRepository,
		resourceRepository: resourceRepository,
	}, nil
}

func (v *VideoValidator) ValidateGetRequestDTO(req dtointerface.GetVideoRequest) error {
	if req.GetID().Value.IsZero() {
		return errtype.NewFieldCannotBeEmptyError(idField)
	}
	if req.GetUserID().Value.IsZero() {
		return errtype.NewFieldCannotBeEmptyError(userIDField)
	}
	return nil
}

func (v *VideoValidator) ValidateListRequestDTO(req dtointerface.ListVideoRequest) error {
	if req.GetName() != "" && len(req.GetName()) <= 3 {
		return errtype.NewFieldLengthMustBeMoreOrLessError(nameField, true, 3)
	}
	if !req.GetCreatedAt().IsZero() && (!req.GetFrom().IsZero() || !req.GetTo().IsZero()) {
		return errtype.NewInternalValidationError("field 'from' or 'to' cannot be passed with 'createdAt'")
	}
	return nil
}

func (v *VideoValidator) ValidateCreateRequestDTO(req dtointerface.CreateVideoRequest) error {
	if req.GetUserID().Value.IsZero() {
		return errtype.NewFieldCannotBeEmptyError(userIDField)
	}
	if req.GetName() == "" {
		return errtype.NewFieldCannotBeEmptyError(nameField)
	}
	if req.GetResourceID().Value.IsZero() {
		return errtype.NewFieldCannotBeEmptyError(resourceIDField)
	}
	return nil
}

func (v *VideoValidator) ValidateUpdateRequestDTO(req dtointerface.UpdateVideoRequest) error {
	if err := v.ValidateGetRequestDTO(req); err != nil {
		return err
	}
	return nil
}

func (v *VideoValidator) ValidateDeleteRequestDTO(req dtointerface.DeleteVideoRequest) error {
	return v.ValidateGetRequestDTO(req)
}

func (v *VideoValidator) ValidateAggregate(agg *agg.Video) error {
	// video fields validation
	if agg.Name == "" {
		return errtype.NewInternalValidationError("'name' cannot be empty")
	}
	if agg.Resource.ID.Value.IsZero() {
		return errtype.NewInternalValidationError("'resource.id' cannot be empty")
	}
	if agg.UserID.Value.IsZero() {
		return errtype.NewInternalValidationError("'userID' cannot be empty")
	}

	// resource fields validation
	if err := v.resourceValidator.ValidateEntity(agg.Resource); err != nil {
		return err
	}

	// video validation by name which must be unique
	q := dto.NewVideoGetRequestDTO(vo.ID{}, agg.Name, vo.ID{}, agg.UserID)
	video, err := v.videoRepository.FindOneByName(v.ctx, q)
	if err != nil {
		if !errtype.IsEntityNotFoundError(err) {
			return v.logger.LogPropagate(err)
		}
	} else {
		if !agg.ID.Value.IsZero() {
			if video.ID.Value != agg.ID.Value {
				return v.logger.LogPropagate(errtype.NewUniquenessCheckFailedError(nameField))
			}
		} else {
			return v.logger.LogPropagate(errtype.NewUniquenessCheckFailedError(nameField))
		}
	}

	// video validation by resource.id which must be unique too
	q = dto.NewVideoGetRequestDTO(vo.ID{}, "", agg.Resource.ID, agg.UserID)
	video, err = v.videoRepository.FindOneByResourceID(v.ctx, q)
	if err != nil {
		if !errtype.IsEntityNotFoundError(err) {
			return v.logger.LogPropagate(err)
		}
	} else {
		if !agg.ID.Value.IsZero() {
			if video.ID.Value != agg.ID.Value {
				return v.logger.LogPropagate(errtype.NewUniquenessCheckFailedError(resourceIDField))
			}
		} else {
			return v.logger.LogPropagate(errtype.NewUniquenessCheckFailedError(resourceIDField))
		}
	}

	return nil
}
