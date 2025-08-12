package iface

type Handler interface {
	HandlePayload(id string, p []byte, progressChan chan<- ProgressInfo) (any, error)
}

type HandlerFunc func(id string, p []byte, progressChan chan<- ProgressInfo) (any, error)

func (f HandlerFunc) HandlePayload(id string, p []byte, progressChan chan<- ProgressInfo) (any, error) {
	return f(id, p, progressChan)
}

type ProgressInfo struct {
	NodeID     string `json:"node_id"`
	PercentNum int    `json:"percent_num"`
	Hostname   string `json:"hostname"`
}
