package socket

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/up9inc/mizu/cli/logger"
	"github.com/up9inc/mizu/cli/uiUtils"
	"github.com/up9inc/mizu/shared"
)

var toastMessageChannel chan *shared.ToastMessage = make(chan *shared.ToastMessage)

func OpenWebsocket(urlStr string) {
	u, err := url.Parse(urlStr)
	if err != nil {
		logger.Log.Errorf("WebSocket URL parse error:", err)
	}
	u.Scheme = "ws"
	u.Path = fmt.Sprintf("%s/ws", u.Path)

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		logger.Log.Errorf("WebSocket Dial error:", err)
	}
	defer conn.Close()

	for {
		msg := <-toastMessageChannel
		wsMsg := &shared.WebSocketToastMessage{
			WebSocketMessageMetadata: &shared.WebSocketMessageMetadata{
				MessageType: shared.WebSocketMessageTypeToast,
			},
			Data: msg,
		}
		data, err := json.Marshal(wsMsg)
		if err != nil {
			logger.Log.Errorf("WebSocket JSON Marshal error:", err)
		}
		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			logger.Log.Errorf("WebSocket write error:", err)
		}
	}
}

func Send(_type string, autoClose uint, text string, metaname string, appendMetaname bool) {
	if appendMetaname {
		text = fmt.Sprintf("%s [%s]", text, metaname)
	}

	switch _type {
	case "success":
		logger.Log.Infof(uiUtils.Green, text)
	case "warning":
		logger.Log.Infof(uiUtils.Yellow, text)
	case "error":
		logger.Log.Infof(uiUtils.Red, text)
	default:
		logger.Log.Infof(uiUtils.White, text)
	}

	toastMessageChannel <- &shared.ToastMessage{
		Type:      _type,
		AutoClose: autoClose,
		Text:      text,
	}
}
