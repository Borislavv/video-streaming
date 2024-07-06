package cache

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Borislavv/video-streaming/internal/domain/agg"
	"github.com/Borislavv/video-streaming/internal/domain/errtype"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	cacherinterface "github.com/Borislavv/video-streaming/internal/domain/service/cacher/interface"
	diinterface "github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	"github.com/Borislavv/video-streaming/internal/infrastructure/helper"
	queryinterface "github.com/Borislavv/video-streaming/internal/infrastructure/repository/query/interface"
	"github.com/Borislavv/video-streaming/internal/infrastructure/repository/storage/mongodb"
	mongodbinterface "github.com/Borislavv/video-streaming/internal/infrastructure/repository/storage/mongodb/interface"
	"reflect"
	"time"
)

var (
	VideoNotFoundByIdError         = errtype.NewEntityNotFoundError("cache", "video", "id")
	VideoNotFoundByNameError       = errtype.NewEntityNotFoundError("cache", "video", "name")
	VideoNotFoundByResourceIdError = errtype.NewEntityNotFoundError("cache", "video", "resource.id")
)

type VideoRepository struct {
	mongodbinterface.Video
	logger loggerinterface.Logger
	cache  cacherinterface.Cacher
}

func NewVideoRepository(serviceContainer diinterface.ServiceContainer) (*VideoRepository, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	videoMongoDbRepository, err := serviceContainer.GetVideoMongoRepository()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	cacheService, err := serviceContainer.GetCacheService()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &VideoRepository{
		cache:  cacheService,
		logger: loggerService,
		Video:  videoMongoDbRepository,
	}, nil
}

func (r *VideoRepository) FindOneByID(ctx context.Context, q queryinterface.FindOneVideoByID) (*agg.Video, error) {
	// attempt to fetch data from cache
	if video, err := r.findOneByID(ctx, q); err == nil {
		return video, nil
	}
	// fetch data from storage if an error occurred
	return r.Video.FindOneByID(ctx, q)
}

func (r *VideoRepository) findOneByID(ctx context.Context, q queryinterface.FindOneVideoByID) (*agg.Video, error) {
	p, err := json.Marshal(q)
	if err != nil {
		return nil, r.logger.LogPropagate(err)
	}
	cacheKey := helper.MD5(p)

	// fetching data from cache/storage
	videoInterface, err := r.cache.Get(
		cacheKey,
		func(item cacherinterface.CacheItem) (data interface{}, err error) {
			item.SetTTL(time.Hour)

			videoAgg, err := r.Video.FindOneByID(ctx, q)
			if err != nil {
				if errors.Is(err, mongodb.VideoNotFoundByIdError) {
					return nil, VideoNotFoundByIdError
				}
				return nil, r.logger.LogPropagate(err)
			}
			return videoAgg, nil
		})
	if err != nil {
		if errors.Is(err, mongodb.VideoNotFoundByIdError) {
			return nil, VideoNotFoundByIdError
		} else {
			return nil, r.logger.LogPropagate(err)
		}
	}

	// casting found data to struct
	videoAgg, ok := videoInterface.(*agg.Video)
	if !ok {
		return nil, errtype.NewCachedDataTypeWasNotMatchedError(
			cacheKey, reflect.TypeOf(&agg.User{}), reflect.TypeOf(videoInterface),
		)
	}

	return videoAgg, nil
}

func (r *VideoRepository) FindList(ctx context.Context, q queryinterface.FindVideoList) (list []*agg.Video, total int64, err error) {
	// attempt to fetch data from cache
	if list, total, err = r.findList(ctx, q); err == nil {
		return list, total, nil
	}
	// fetch data from storage if an error occurred
	return r.Video.FindList(ctx, q)
}

func (r *VideoRepository) findList(ctx context.Context, q queryinterface.FindVideoList) (list []*agg.Video, total int64, err error) {
	p, err := json.Marshal(q)
	if err != nil {
		return nil, 0, r.logger.LogPropagate(err)
	}
	cacheKey := helper.MD5(p)

	type response struct {
		List  []*agg.Video
		Total int64
	}

	responseInterface, err := r.cache.Get(
		cacheKey,
		func(item cacherinterface.CacheItem) (data interface{}, err error) {
			item.SetTTL(time.Hour)

			l, t, e := r.Video.FindList(ctx, q)
			if err != nil {
				return nil, r.logger.LogPropagate(e)
			}

			return response{List: l, Total: t}, nil
		},
	)
	if err != nil {
		return nil, 0, r.logger.LogPropagate(err)
	}

	listResponse, ok := responseInterface.(response)
	if !ok {
		return nil, 0, errtype.NewCachedDataTypeWasNotMatchedError(
			cacheKey, reflect.TypeOf(response{}), reflect.TypeOf(responseInterface),
		)
	}

	return listResponse.List, listResponse.Total, nil
}

func (r *VideoRepository) FindOneByName(ctx context.Context, q queryinterface.FindOneVideoByName) (*agg.Video, error) {
	// attempt to fetch data from cache
	if video, err := r.findOneByName(ctx, q); err == nil {
		return video, nil
	}
	// fetch data from storage if an error occurred
	return r.Video.FindOneByName(ctx, q)
}

func (r *VideoRepository) findOneByName(ctx context.Context, q queryinterface.FindOneVideoByName) (*agg.Video, error) {
	p, err := json.Marshal(q)
	if err != nil {
		return nil, r.logger.LogPropagate(err)
	}
	cacheKey := helper.MD5(p)

	videoInterface, err := r.cache.Get(cacheKey, func(item cacherinterface.CacheItem) (data interface{}, err error) {
		item.SetTTL(time.Hour)

		videoAgg, err := r.Video.FindOneByName(ctx, q)
		if err != nil {
			if errors.Is(err, mongodb.VideoNotFoundByNameError) {
				return nil, VideoNotFoundByNameError
			}
			return nil, r.logger.LogPropagate(err)
		}

		return videoAgg, nil
	})
	if err != nil {
		if errors.Is(err, VideoNotFoundByNameError) {
			return nil, VideoNotFoundByNameError
		} else {
			return nil, r.logger.LogPropagate(err)
		}
	}

	videoAgg, ok := videoInterface.(*agg.Video)
	if !ok {
		return nil, errtype.NewCachedDataTypeWasNotMatchedError(
			cacheKey, reflect.TypeOf(&agg.Video{}), reflect.TypeOf(videoInterface),
		)
	}

	return videoAgg, nil
}

func (r *VideoRepository) FindOneByResourceID(ctx context.Context, q queryinterface.FindOneVideoByResourceID) (*agg.Video, error) {
	// attempt to fetch data from cache
	if video, err := r.findOneByResourceID(ctx, q); err == nil {
		return video, nil
	}
	// fetch data from storage if an error occurred
	return r.Video.FindOneByResourceID(ctx, q)
}

func (r *VideoRepository) findOneByResourceID(ctx context.Context, q queryinterface.FindOneVideoByResourceID) (*agg.Video, error) {
	p, err := json.Marshal(q)
	if err != nil {
		return nil, r.logger.LogPropagate(err)
	}
	cacheKey := helper.MD5(p)

	videoInterface, err := r.cache.Get(cacheKey, func(item cacherinterface.CacheItem) (data interface{}, err error) {
		item.SetTTL(time.Hour)

		videoAgg, err := r.Video.FindOneByResourceID(ctx, q)
		if err != nil {
			if errors.Is(err, mongodb.VideoNotFoundByResourceIdError) {
				return nil, VideoNotFoundByResourceIdError
			}
			return nil, r.logger.LogPropagate(err)
		}

		return videoAgg, nil
	})
	if err != nil {
		if errors.Is(err, VideoNotFoundByResourceIdError) {
			return nil, VideoNotFoundByResourceIdError
		} else {
			return nil, r.logger.LogPropagate(err)
		}
	}

	videoAgg, ok := videoInterface.(*agg.Video)
	if !ok {
		return nil, errtype.NewCachedDataTypeWasNotMatchedError(
			cacheKey, reflect.TypeOf(&agg.Video{}), reflect.TypeOf(videoInterface),
		)
	}

	return videoAgg, nil
}
