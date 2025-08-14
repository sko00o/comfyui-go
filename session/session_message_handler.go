package session

import (
	"fmt"
	"io"

	"github.com/gorilla/websocket"

	"github.com/sko00o/comfyui-go/helper"
	"github.com/sko00o/comfyui-go/iface"
)

var _ iface.MessageHandler = (*WrapSession)(nil)

type WrapSession struct {
	*Session
}

func (s *WrapSession) WriteMessage(messageType int, data []byte) error {
	switch messageType {
	case websocket.TextMessage:
		s.Logger.Debugf("ws recv TXT: %s", data)
		s.handleTextMessage(data)
	case websocket.BinaryMessage:
		s.Logger.Debugf("ws recv BIN: %x...", helper.Head(data))
		s.handleBinaryMessage(data)
	default:
		return fmt.Errorf("unsupported message type: %d", messageType)
	}
	return nil
}

func (s *WrapSession) ReadMessage() (messageType int, p []byte, err error) {
	<-s.done
	return 0, nil, io.EOF
}

func (s *WrapSession) Name() string {
	return fmt.Sprintf("comfy:%s", s.TaskID)
}
