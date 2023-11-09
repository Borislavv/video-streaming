package builder

import (
	"context"
	"encoding/json"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/dto"
	"github.com/Borislavv/video-streaming/internal/domain/entity"
	"github.com/Borislavv/video-streaming/internal/domain/enum"
	"github.com/Borislavv/video-streaming/internal/domain/logger"
	"github.com/Borislavv/video-streaming/internal/domain/repository"
	"github.com/Borislavv/video-streaming/internal/domain/service/extractor"
	"github.com/Borislavv/video-streaming/internal/domain/vo"
	"github.com/Borislavv/video-streaming/internal/infrastructure/helper"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"net/http"
	"strconv"
	"time"
)

const (
	idField           = "id"
	nameField         = "name"
	createdAtField    = "createdAt"
	fromField         = "from"
	toField           = "to"
	pageField         = "page"
	limitField        = "limit"
	limitDefaultValue = 25
	pageDefaultValue  = 1
)

type VideoBuilder struct {
	logger             logger.Logger
	ctx                context.Context
	extractor          extractor.RequestParams
	videoRepository    repository.Video
	resourceRepository repository.Resource
}

// NewVideoBuilder is a constructor of VideoBuilder
func NewVideoBuilder(
	ctx context.Context,
	logger logger.Logger,
	extractor extractor.RequestParams,
	videoRepository repository.Video,
	resourceRepository repository.Resource,
) *VideoBuilder {
	return &VideoBuilder{
		ctx:                ctx,
		logger:             logger,
		extractor:          extractor,
		videoRepository:    videoRepository,
		resourceRepository: resourceRepository,
	}
}

// BuildCreateRequestDTOFromRequest - build a dto.CreateVideoRequest from raw *http.Request
func (b *VideoBuilder) BuildCreateRequestDTOFromRequest(r *http.Request) (*dto.VideoCreateRequestDTO, error) {
	videoDTO := &dto.VideoCreateRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(videoDTO); err != nil {
		return nil, b.logger.LogPropagate(err)
	}

	if userID, ok := r.Context().Value(enum.UserIDContextKey).(vo.ID); ok {
		videoDTO.UserID = userID
	}

	return videoDTO, nil
}

// BuildAggFromCreateRequestDTO - build an agg.Video from dto.CreateVideoRequest
func (b *VideoBuilder) BuildAggFromCreateRequestDTO(reqDTO dto.CreateVideoRequest) (*agg.Video, error) {
	resource, err := b.resourceRepository.Find(b.ctx, reqDTO.GetResourceID())
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}

	return &agg.Video{
		Video: entity.Video{
			UserID:      reqDTO.GetUserID(),
			Name:        reqDTO.GetName(),
			Description: reqDTO.GetDescription(),
		},
		Resource: resource.Resource,
		Timestamp: vo.Timestamp{
			CreatedAt: time.Now(),
		},
	}, nil
}

// BuildUpdateRequestDTOFromRequest - build a dto.UpdateVideoRequest from raw *http.Request
func (b *VideoBuilder) BuildUpdateRequestDTOFromRequest(r *http.Request) (*dto.VideoUpdateRequestDTO, error) {
	videoDTO := &dto.VideoUpdateRequestDTO{}
	if err := json.NewDecoder(r.Body).Decode(&videoDTO); err != nil {
		return nil, b.logger.LogPropagate(err)
	}

	// setting up a user id
	if userID, ok := r.Context().Value(enum.UserIDContextKey).(vo.ID); ok {
		videoDTO.UserID = userID
	}

	// setting up a video id
	hexID, err := b.extractor.GetParameter(idField, r)
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}
	oID, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}
	videoDTO.ID = vo.ID{Value: oID}

	return videoDTO, nil
}

// BuildAggFromUpdateRequestDTO - build an agg.Video from dto.UpdateVideoRequest
func (b *VideoBuilder) BuildAggFromUpdateRequestDTO(reqDTO dto.UpdateVideoRequest) (*agg.Video, error) {
	video, err := b.videoRepository.Find(b.ctx, reqDTO.GetID())
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}

	changes := 0
	if video.Name != reqDTO.GetName() {
		video.Name = reqDTO.GetName()
		changes++
	}
	if video.Description != reqDTO.GetDescription() {
		video.Description = reqDTO.GetDescription()
		changes++
	}
	if !reqDTO.GetResourceID().Value.IsZero() {
		resource, ferr := b.resourceRepository.Find(b.ctx, reqDTO.GetResourceID())
		if ferr != nil {
			return nil, b.logger.LogPropagate(ferr)
		}
		if video.Resource.ID.Value != resource.Resource.ID.Value {
			video.Resource = resource.Resource
			changes++
		}
	}
	if changes > 0 {
		video.Timestamp.UpdatedAt = time.Now()
	}

	return video, nil
}

// BuildGetRequestDTOFromRequest - build a dto.GetVideoRequest from raw *http.Request
func (b *VideoBuilder) BuildGetRequestDTOFromRequest(r *http.Request) (*dto.VideoGetRequestDTO, error) {
	videoDTO := &dto.VideoGetRequestDTO{}

	hexID, err := b.extractor.GetParameter(idField, r)
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}
	oID, err := primitive.ObjectIDFromHex(hexID)
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}
	videoDTO.ID = vo.ID{Value: oID}

	return videoDTO, nil
}

// BuildListRequestDTOFromRequest - build a dto.ListVideoRequest from raw *http.Request
func (b *VideoBuilder) BuildListRequestDTOFromRequest(r *http.Request) (*dto.VideoListRequestDTO, error) {
	videoDTO := &dto.VideoListRequestDTO{}

	if b.extractor.HasParameter(nameField, r) {
		if nm, err := b.extractor.GetParameter(nameField, r); err == nil {
			videoDTO.Name = nm
		}
	}
	if b.extractor.HasParameter(createdAtField, r) {
		createdAt, _ := b.extractor.GetParameter(createdAtField, r)

		parsedCreatedAt, err := helper.ParseTime(createdAt)
		if err != nil {
			return nil, b.logger.LogPropagate(err)
		} else {
			videoDTO.CreatedAt = parsedCreatedAt
		}
	}
	if b.extractor.HasParameter(fromField, r) {
		from, _ := b.extractor.GetParameter(fromField, r)

		parsedFrom, err := helper.ParseTime(from)
		if err != nil {
			return nil, b.logger.LogPropagate(err)
		} else {
			videoDTO.From = parsedFrom
		}
	}
	if b.extractor.HasParameter(toField, r) {
		to, _ := b.extractor.GetParameter(toField, r)

		parsedTo, err := helper.ParseTime(to)
		if err != nil {
			return nil, b.logger.LogPropagate(err)
		} else {
			videoDTO.To = parsedTo
		}
	}
	if b.extractor.HasParameter(pageField, r) {
		pg, _ := b.extractor.GetParameter(pageField, r)
		pgi, atoiErr := strconv.Atoi(pg)
		if atoiErr != nil {
			return nil, b.logger.LogPropagate(atoiErr)
		}
		videoDTO.Page = pgi
	} else {
		videoDTO.Page = pageDefaultValue
	}
	if b.extractor.HasParameter(limitField, r) {
		l, _ := b.extractor.GetParameter(limitField, r)
		li, atoiErr := strconv.Atoi(l)
		if atoiErr != nil {
			return nil, b.logger.LogPropagate(atoiErr)
		}
		videoDTO.Limit = li
	} else {
		videoDTO.Limit = limitDefaultValue
	}

	return videoDTO, nil
}

// BuildDeleteRequestDTOFromRequest - build a dto.DeleteVideoRequest from raw *http.Request
func (b *VideoBuilder) BuildDeleteRequestDTOFromRequest(r *http.Request) (*dto.VideoDeleteRequestDto, error) {
	videoGetDTO, err := b.BuildGetRequestDTOFromRequest(r)
	if err != nil {
		return nil, b.logger.LogPropagate(err)
	}

	return &dto.VideoDeleteRequestDto{ID: videoGetDTO.ID}, nil
}
