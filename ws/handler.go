package ws

type Handler interface {
	HandleMessage(messageType int, message []byte)
}

type HandlerFunc func(int, []byte)

func (f HandlerFunc) HandleMessage(messageType int, message []byte) {
	f(messageType, message)
}
