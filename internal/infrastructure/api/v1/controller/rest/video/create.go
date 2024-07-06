package video

import (
	builderinterface "github.com/Borislavv/video-streaming/internal/domain/builder/interface"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	authenticatorinterface "github.com/Borislavv/video-streaming/internal/domain/service/authenticator/interface"
	diinterface "github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	videointerface "github.com/Borislavv/video-streaming/internal/domain/service/video/interface"
	responseinterface "github.com/Borislavv/video-streaming/internal/infrastructure/api/v1/response/interface"
	"github.com/gorilla/mux"
	"net/http"
)

const CreatePath = "/video"

type CreateController struct {
	logger      loggerinterface.Logger
	builder     builderinterface.Video
	service     videointerface.CRUD
	authService authenticatorinterface.Authenticator
	responder   responseinterface.Responder
}

func NewCreateController(serviceContainer diinterface.ServiceContainer) (*CreateController, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	videoBuilder, err := serviceContainer.GetVideoBuilder()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	videoCRUDService, err := serviceContainer.GetVideoCRUDService()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	authService, err := serviceContainer.GetAuthService()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	responseService, err := serviceContainer.GetResponderService()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &CreateController{
		logger:      loggerService,
		builder:     videoBuilder,
		service:     videoCRUDService,
		authService: authService,
		responder:   responseService,
	}, nil
}

func (c *CreateController) Create(w http.ResponseWriter, r *http.Request) {
	videoDTO, err := c.builder.BuildCreateRequestDTOFromRequest(r)
	if err != nil {
		c.responder.Respond(w, c.logger.LogPropagate(err))
		return
	}

	videoAgg, err := c.service.Create(videoDTO)
	if err != nil {
		c.responder.Respond(w, c.logger.LogPropagate(err))
		return
	}

	w.WriteHeader(http.StatusCreated)
	c.responder.Respond(w, videoAgg)
}

func (c *CreateController) AddRoute(router *mux.Router) {
	router.
		Path(CreatePath).
		HandlerFunc(c.Create).
		Methods(http.MethodPost)
}
