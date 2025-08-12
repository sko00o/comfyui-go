package prompt

import (
	"encoding/json"

	"github.com/sko00o/comfyui-go/node"
)

type Builder interface {
	IBuilder
	OutputNodeIDs() []string
}

type IBuilder interface {
	Build() Prompt
}

type Prompt map[string]node.Builder

func (w Prompt) MarshalJSON() ([]byte, error) {
	workflow := make(map[string]node.Node, len(w))
	for id, nb := range w {
		workflow[id] = nb.Build()
	}
	return json.Marshal(workflow)
}
