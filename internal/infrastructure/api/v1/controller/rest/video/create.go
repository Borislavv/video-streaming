package video

import (
	"github.com/Borislavv/video-streaming/internal/domain/builder"
	"github.com/Borislavv/video-streaming/internal/domain/logger"
	"github.com/Borislavv/video-streaming/internal/domain/service"
	"github.com/Borislavv/video-streaming/internal/infrastructure/api/v1/response"
	"github.com/gorilla/mux"
	"net/http"
)

const CreatePath = "/video"

type CreateVideoController struct {
	logger   logger.Logger
	builder  builder.Video
	service  service.Video
	response response.Responder
}

func NewCreateController(
	logger logger.Logger,
	builder builder.Video,
	service service.Video,
	response response.Responder,
) *CreateVideoController {
	return &CreateVideoController{
		logger:   logger,
		builder:  builder,
		service:  service,
		response: response,
	}
}

func (c *CreateVideoController) Create(w http.ResponseWriter, r *http.Request) {
	dto, err := c.builder.BuildCreateRequestDTOFromRequest(r)
	if err != nil {
		c.response.Respond(w, c.logger.LogPropagate(err))
		return
	}

	video, err := c.service.Create(dto)
	if err != nil {
		c.response.Respond(w, c.logger.LogPropagate(err))
		return
	}

	c.response.Respond(w, video)
	w.WriteHeader(http.StatusCreated)
}

func (c *CreateVideoController) AddRoute(router *mux.Router) {
	router.
		Path(CreatePath).
		HandlerFunc(c.Create).
		Methods(http.MethodPost)
}
