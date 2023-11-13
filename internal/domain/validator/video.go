package validator

import (
	"context"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/dto"
	"github.com/Borislavv/video-streaming/internal/domain/errors"
	"github.com/Borislavv/video-streaming/internal/domain/logger"
	"github.com/Borislavv/video-streaming/internal/domain/repository"
	"github.com/Borislavv/video-streaming/internal/domain/service/accessor"
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
	logger             logger.Logger
	resourceValidator  Resource
	accessService      accessor.Accessor
	videoRepository    repository.Video
	resourceRepository repository.Resource
}

func NewVideoValidator(
	ctx context.Context,
	logger logger.Logger,
	resourceValidator Resource,
	accessService accessor.Accessor,
	videoRepository repository.Video,
	resourceRepository repository.Resource,
) *VideoValidator {
	return &VideoValidator{
		ctx:                ctx,
		logger:             logger,
		resourceValidator:  resourceValidator,
		accessService:      accessService,
		videoRepository:    videoRepository,
		resourceRepository: resourceRepository,
	}
}

func (v *VideoValidator) ValidateGetRequestDTO(req dto.GetVideoRequest) error {
	if req.GetID().Value.IsZero() {
		return errors.NewFieldCannotBeEmptyError(idField)
	}
	if req.GetUserID().Value.IsZero() {
		return errors.NewFieldCannotBeEmptyError(userIDField)
	}
	return nil
}

func (v *VideoValidator) ValidateListRequestDTO(req dto.ListVideoRequest) error {
	if req.GetName() != "" && len(req.GetName()) <= 3 {
		return errors.NewFieldLengthMustBeMoreOrLessError(nameField, true, 3)
	}
	if !req.GetCreatedAt().IsZero() && (!req.GetFrom().IsZero() || !req.GetTo().IsZero()) {
		return errors.NewInternalValidationError("field 'from' or 'to' cannot be passed with 'createdAt'")
	}
	return nil
}

func (v *VideoValidator) ValidateCreateRequestDTO(req dto.CreateVideoRequest) error {
	if req.GetUserID().Value.IsZero() {
		return errors.NewFieldCannotBeEmptyError(userIDField)
	}
	if req.GetName() == "" {
		return errors.NewFieldCannotBeEmptyError(nameField)
	}
	if req.GetResourceID().Value.IsZero() {
		return errors.NewFieldCannotBeEmptyError(resourceIDField)
	}
	return nil
}

func (v *VideoValidator) ValidateUpdateRequestDTO(req dto.UpdateVideoRequest) error {
	if err := v.ValidateGetRequestDTO(req); err != nil {
		return err
	}
	return nil
}

func (v *VideoValidator) ValidateDeleteRequestDTO(req dto.DeleteVideoRequest) error {
	return v.ValidateGetRequestDTO(req)
}

func (v *VideoValidator) ValidateAggregate(agg *agg.Video) error {
	// video fields validation
	if agg.Name == "" {
		return errors.NewInternalValidationError("'name' cannot be empty")
	}
	if agg.Resource.ID.Value.IsZero() {
		return errors.NewInternalValidationError("'resource.id' cannot be empty")
	}
	if agg.UserID.Value.IsZero() {
		return errors.NewInternalValidationError("'userID' cannot be empty")
	}

	// resource fields validation
	if err := v.resourceValidator.ValidateEntity(agg.Resource); err != nil {
		return err
	}

	// video validation by name which must be unique
	q := dto.NewVideoGetRequestDTO(vo.ID{}, agg.Name, vo.ID{}, agg.UserID)
	video, err := v.videoRepository.FindOneByName(v.ctx, q)
	if err != nil {
		if !errors.IsEntityNotFoundError(err) {
			return v.logger.LogPropagate(err)
		}
	} else {
		if !agg.ID.Value.IsZero() {
			if video.ID.Value != agg.ID.Value {
				return errors.NewUniquenessCheckFailedError(nameField)
			}
		} else {
			return errors.NewUniquenessCheckFailedError(nameField)
		}
	}

	// video validation by resource.id which must be unique too
	q = dto.NewVideoGetRequestDTO(vo.ID{}, "", agg.Resource.ID, agg.UserID)
	video, err = v.videoRepository.FindOneByResourceID(v.ctx, q)
	if err != nil {
		if !errors.IsEntityNotFoundError(err) {
			return v.logger.LogPropagate(err)
		}
	} else {
		if !agg.ID.Value.IsZero() {
			if video.ID.Value != agg.ID.Value {
				return errors.NewUniquenessCheckFailedError(resourceIDField)
			}
		} else {
			return errors.NewUniquenessCheckFailedError(resourceIDField)
		}
	}

	return nil
}
