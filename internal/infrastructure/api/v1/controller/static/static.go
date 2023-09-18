package static

import (
	"github.com/Borislavv/video-streaming/internal/domain/logger"
	"github.com/Borislavv/video-streaming/internal/infrastructure/api/v1/response"
	"github.com/Borislavv/video-streaming/internal/infrastructure/helper"
	"github.com/gorilla/mux"
	"net/http"
	"strings"
)

const ResourcesPrefix = "/static/"

type ResourceController struct {
	logger    logger.Logger
	responder response.Responder
}

func NewResourceController(
	logger logger.Logger,
	responder response.Responder,
) *ResourceController {
	return &ResourceController{
		logger:    logger,
		responder: responder,
	}
}

func (c *ResourceController) Serve(w http.ResponseWriter, r *http.Request) {
	dir, err := helper.StaticFilesDir()
	if err != nil {
		c.responder.Respond(w, c.logger.LogPropagate(err))
		return
	}

	path := r.URL.Path
	if strings.Contains(path, ResourcesPrefix) {
		path = strings.ReplaceAll(path, ResourcesPrefix, "")
	}

	http.ServeFile(w, r, dir+path)
}

func (c *ResourceController) AddRoute(router *mux.Router) {
	router.
		PathPrefix(ResourcesPrefix).
		HandlerFunc(c.Serve).
		Methods(http.MethodGet)
}
