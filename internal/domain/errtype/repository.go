package errtype

import (
	"fmt"
	"github.com/Borislavv/video-streaming/internal/infrastructure/service/logger"
	"net/http"
)

const (
	repositoryType = "application"

	publicRepositoryLevel  = logger.ErrorLevel
	publicRepositoryStatus = http.StatusNotFound

	internalRepositoryLevel  = logger.CriticalLevel
	internalRepositoryStatus = http.StatusInternalServerError
)

type EntityNotFoundError struct{ publicError }

func NewEntityNotFoundError(where, entity string, by string) *EntityNotFoundError {
	return &EntityNotFoundError{
		publicError{
			errored{
				ErrorMessage: fmt.Sprintf("%v: '%v' not found by given '%s'", where, entity, by),
				ErrorType:    repositoryType,
				errorLevel:   publicRepositoryLevel,
				errorStatus:  publicRepositoryStatus,
			},
		},
	}
}

func IsEntityNotFoundError(err error) bool {
	_, ok := err.(*EntityNotFoundError)
	return ok
}

type InternalRepositoryError struct{ internalError }

func NewInternalRepositoryError(msg string) *InternalRepositoryError {
	return &InternalRepositoryError{
		internalError{
			errored{
				ErrorMessage: msg,
				ErrorType:    repositoryType,
				errorLevel:   internalRepositoryLevel,
				errorStatus:  internalRepositoryStatus,
			},
		},
	}
}
