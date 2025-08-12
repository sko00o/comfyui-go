package error

import (
	"encoding/json"
	"time"
)

type ComfyUIError struct {
	Message   json.RawMessage
	IsOOM     bool
	NodesTime map[string]time.Duration
}

func (e ComfyUIError) Error() string {
	return string(e.Message)
}
