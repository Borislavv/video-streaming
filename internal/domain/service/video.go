package service

import (
	"context"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/builder"
	"github.com/Borislavv/video-streaming/internal/domain/dto"
	"github.com/Borislavv/video-streaming/internal/domain/logger"
	"github.com/Borislavv/video-streaming/internal/domain/repository"
	"github.com/Borislavv/video-streaming/internal/domain/validator"
)

type VideoService struct {
	ctx        context.Context
	logger     logger.Logger
	builder    builder.Video
	validator  validator.Video
	repository repository.Video
}

func NewVideoService(
	ctx context.Context,
	logger logger.Logger,
	builder builder.Video,
	validator validator.Video,
	repository repository.Video,
) *VideoService {
	return &VideoService{
		ctx:        ctx,
		logger:     logger,
		builder:    builder,
		validator:  validator,
		repository: repository,
	}
}

func (s *VideoService) Get(req dto.GetRequest) (*agg.Video, error) {
	// validation of input request
	if err := s.validator.ValidateGetRequestDTO(req); err != nil {
		return nil, err
	}

	video, err := s.repository.Find(s.ctx, req.GetId())
	if err != nil {
		return nil, s.logger.ErrorPropagate(err)
	}

	return video, nil
}

func (s *VideoService) List(req dto.ListRequest) (list []*agg.Video, total int64, err error) {
	// validation of input request
	if err = s.validator.ValidateListRequestDTO(req); err != nil {
		return nil, 0, s.logger.LogPropagate(err)
	}

	list, total, err = s.repository.FindList(s.ctx, req)
	if err != nil {
		return nil, 0, s.logger.LogPropagate(err)
	}

	return list, total, err
}

func (s *VideoService) Create(req dto.CreateRequest) (*agg.Video, error) {
	// validation of input request
	if err := s.validator.ValidateCreateRequestDTO(req); err != nil {
		return nil, err
	}

	// building an aggregate
	video, err := s.builder.BuildAggFromCreateRequestDTO(req)
	if err != nil {
		return nil, s.logger.LogPropagate(err)
	}

	// validation of an aggregate
	if err = s.validator.ValidateAggregate(video); err != nil {
		return nil, s.logger.ErrorPropagate(err)
	}

	// saving an aggregate into storage
	video, err = s.repository.Insert(s.ctx, video)
	if err != nil {
		return nil, s.logger.ErrorPropagate(err)
	}

	return video, nil
}

func (s *VideoService) Update(req dto.UpdateRequest) (*agg.Video, error) {
	// validation of input request
	if err := s.validator.ValidateUpdateRequestDTO(req); err != nil {
		return nil, s.logger.LogPropagate(err)
	}

	// building an aggregate
	videoAgg, err := s.builder.BuildAggFromUpdateRequestDTO(req)
	if err != nil {
		return nil, s.logger.LogPropagate(err)
	}

	// validation of an aggregate
	if err = s.validator.ValidateAggregate(videoAgg); err != nil {
		return nil, s.logger.LogPropagate(err)
	}

	// saving updated aggregate into storage
	videoAgg, err = s.repository.Update(s.ctx, videoAgg)
	if err != nil {
		return nil, s.logger.LogPropagate(err)
	}

	return videoAgg, nil
}

func (s *VideoService) Delete(req dto.DeleteRequest) error {
	// validation of input request
	if err := s.validator.ValidateDeleteRequestDTO(req); err != nil {
		return err
	}

	// fetching a video which will be deleted
	video, err := s.repository.Find(s.ctx, req.GetId())
	if err != nil {
		return s.logger.ErrorPropagate(err)
	}

	// video removing
	if err = s.repository.Remove(s.ctx, video); err != nil {
		return s.logger.ErrorPropagate(err)
	}

	return nil
}
