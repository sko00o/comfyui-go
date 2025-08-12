package node

import (
	"encoding/json"
	"fmt"
)

type Builder interface {
	Build() Node
}

type GeneralNode struct {
	ClassType string
	Inputs    map[string]any
}

func (n GeneralNode) Build() Node {
	return Node{
		ClassType: n.ClassType,
		Inputs:    n.Inputs,
	}
}

type Node struct {
	ClassType string `json:"class_type"`
	Inputs    any    `json:"inputs"`
	Meta      *Meta  `json:"_meta,omitempty"`
}

type Meta struct {
	Title string `json:"title"`
}

type PreNode struct {
	ID   string
	Argc int // which output argument
}

func (n PreNode) MarshalJSON() ([]byte, error) {
	return json.Marshal([]any{n.ID, n.Argc})
}

func (n *PreNode) UnmarshalJSON(p []byte) error {
	m := make([]any, 0, 2)
	if err := json.Unmarshal(p, &m); err != nil {
		return err
	}
	if len(m) < 2 {
		return fmt.Errorf("require 2 fields, got: %s", p)
	}

	switch v := m[0].(type) {
	case string:
		n.ID = v
	default:
		return fmt.Errorf("field 1 has invalid type %v", v)
	}
	switch v := m[1].(type) {
	case float64:
		n.Argc = int(v)
	case int64:
		n.Argc = int(v)
	default:
		return fmt.Errorf("field 2 has invalid type %v", v)
	}
	return nil
}
