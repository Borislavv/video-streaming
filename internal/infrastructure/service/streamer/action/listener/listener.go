package listener

import (
	"fmt"
	"github.com/Borislavv/video-streaming/internal/domain/logger/interface"
	"github.com/Borislavv/video-streaming/internal/domain/service/di/interface"
	"github.com/Borislavv/video-streaming/internal/infrastructure/service/streamer/action/enum"
	"github.com/Borislavv/video-streaming/internal/infrastructure/service/streamer/action/model"
	protointerface "github.com/Borislavv/video-streaming/internal/infrastructure/service/streamer/proto/interface"
	"github.com/gorilla/websocket"
	"sync"
)

var (
	supportedActionsMap = map[enum.Actions]struct{}{
		enum.StreamByID: {},
	}
)

type WebSocketActionsListener struct {
	logger       loggerinterface.Logger
	communicator protointerface.Communicator
}

func NewWebSocketActionsListener(serviceContainer diinterface.ServiceContainer) (*WebSocketActionsListener, error) {
	loggerService, err := serviceContainer.GetLoggerService()
	if err != nil {
		return nil, err
	}

	webSocketCommunicatorService, err := serviceContainer.GetWebSocketCommunicatorService()
	if err != nil {
		return nil, loggerService.LogPropagate(err)
	}

	return &WebSocketActionsListener{
		logger:       loggerService,
		communicator: webSocketCommunicatorService,
	}, nil
}

func (l *WebSocketActionsListener) Listen(wg *sync.WaitGroup, conn *websocket.Conn) <-chan model.Action {
	actionsCh := make(chan model.Action, 1)

	wg.Add(1)
	go func() {
		defer func() {
			close(actionsCh)
			wg.Done()
		}()

		for {
			t, b, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					l.logger.Info(fmt.Sprintf("[%v]: websocket connection has been closed", conn.RemoteAddr()))
					return
				}
				l.logger.Error(fmt.Sprintf("[%v]: %v", conn.RemoteAddr(), err.Error()))
				return
			}
			if t == websocket.TextMessage {
				do, data, err := l.communicator.Parse(b)
				if err != nil {
					l.logger.Error(fmt.Sprintf("[%v]: %v", conn.RemoteAddr(), err.Error()))
					return
				}
				if _, isSupported := supportedActionsMap[do]; isSupported {
					actionsCh <- model.Action{Do: do, Data: data, Conn: conn}
					l.logger.Info(fmt.Sprintf("action '%v' with data '%v' received", do, data))
				} else {
					l.logger.Critical(fmt.Sprintf("do: %+v, data: %+v received unsupport action", do, data))
				}
			}
		}
	}()

	return actionsCh
}
