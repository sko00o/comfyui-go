package message

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ref: https://docs.comfy.org/development/comfyui-server/comms_messages

type Type string

const (
	Status               Type = "status"
	Progress             Type = "progress"
	Executed             Type = "executed"
	Executing            Type = "executing"
	ExecutionStart       Type = "execution_start"
	ExecutionError       Type = "execution_error"
	ExecutionCached      Type = "execution_cached"
	ExecutionSuccess     Type = "execution_success"
	ExecutionInterrupted Type = "execution_interrupted"
)

type Data interface {
	GetPromptID() string
	SetPromptID(promptID string)
}

type Message struct {
	Type Type `json:"type"`
	Data Data `json:"data"`
}

func (m *Message) UnmarshalJSON(b []byte) error {
	type Alias Message
	a := &struct {
		*Alias
		Data json.RawMessage `json:"data"`
	}{
		Alias: (*Alias)(m),
	}
	if err := json.Unmarshal(b, a); err != nil {
		return fmt.Errorf("unmarshal message: %w", err)
	}

	switch m.Type {
	case Status:
		m.Data = &DataStatus{}
	case Executing:
		m.Data = &DataExecuting{}
	case Progress:
		m.Data = &DataProgress{}
	case Executed:
		m.Data = &DataExecuted{}
	case ExecutionStart, ExecutionSuccess, ExecutionCached:
		m.Data = &DataExecution{}
	case ExecutionError:
		m.Data = &DataExecutionError{}
	case ExecutionInterrupted:
		m.Data = &DataExecutionInterrupted{}
	default:
		m.Data = nil // ignore the unknown message data
	}

	if m.Data != nil {
		if err := json.Unmarshal(a.Data, m.Data); err != nil {
			return fmt.Errorf("unmarshal message data: %w", err)
		}
	}
	return nil
}

type DataStatus struct {
	SID    *string `json:"sid"`
	Status struct {
		ExecInfo struct {
			QueueRemaining int `json:"queue_remaining"`
		} `json:"exec_info"`
	} `json:"status"`
}

func (b DataStatus) GetPromptID() string {
	return ""
}

func (b *DataStatus) SetPromptID(promptID string) {}

type ExecutionBase struct {
	PromptID string `json:"prompt_id"`
}

func (b ExecutionBase) GetPromptID() string {
	return b.PromptID
}

func (b *ExecutionBase) SetPromptID(promptID string) {
	b.PromptID = promptID
}

type DataExecuting struct {
	ExecutionBase
	Node        *string `json:"node,omitempty"`
	DisplayNode *string `json:"display_node,omitempty"`
}

type DataProgress struct {
	DataExecuting
	Value int `json:"value"`
	Max   int `json:"max"`
}

func (d *DataProgress) UnmarshalJSON(data []byte) error {
	type Alias DataProgress
	aux := &struct {
		*Alias
		Value any `json:"value"`
		Max   any `json:"max"`
	}{
		Alias: (*Alias)(d),
	}

	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}

	switch v := aux.Value.(type) {
	case float64:
		d.Value = int(v)
	case int:
		d.Value = v
	case int64:
		d.Value = int(v)
	default:
		return fmt.Errorf("value field is not a number: %v", aux.Value)
	}

	switch v := aux.Max.(type) {
	case float64:
		d.Max = int(v)
	case int:
		d.Max = v
	case int64:
		d.Max = int(v)
	default:
		return fmt.Errorf("max field is not a number: %v", aux.Max)
	}

	return nil
}

type DataExecuted struct {
	DataExecuting
	Output MapOutput `json:"output"`
}

type MapOutput map[string]json.RawMessage

type FileInfo struct {
	Filename  string                 `json:"filename"`
	Subfolder string                 `json:"subfolder"`
	Type      string                 `json:"type"`
	Raw       map[string]interface{} `json:"-"`
}

func (f *FileInfo) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &f.Raw); err != nil {
		return err
	}

	if filename, ok := f.Raw["filename"].(string); ok {
		f.Filename = filename
	} else {
		return fmt.Errorf("filename is not a string")
	}
	if subfolder, ok := f.Raw["subfolder"].(string); ok {
		f.Subfolder = subfolder
	} else {
		return fmt.Errorf("subfolder is not a string")
	}
	if typ, ok := f.Raw["type"].(string); ok {
		f.Type = typ
	} else {
		return fmt.Errorf("type is not a string")
	}

	return nil
}

func (f FileInfo) MarshalJSON() ([]byte, error) {
	f.Raw["filename"] = f.Filename
	f.Raw["subfolder"] = f.Subfolder
	f.Raw["type"] = f.Type
	return json.Marshal(f.Raw)
}

type DataExecution struct {
	ExecutionBase
	Nodes     []string `json:"nodes,omitempty"`
	Timestamp *int64   `json:"timestamp,omitempty"`
}

type DataExecutionInterrupted struct {
	ExecutionBase
	NodeID   string   `json:"node_id"`
	NodeType string   `json:"node_type"`
	Executed []string `json:"executed"`
}

const (
	ExceptionTypeOOM = "torch.OutOfMemoryError"
	ExceptionTypeRE  = "RuntimeError"
)

type DataExecutionError struct {
	DataExecutionInterrupted
	ExceptionMessage string          `json:"exception_message"`
	ExceptionType    string          `json:"exception_type"`
	Traceback        json.RawMessage `json:"traceback"`
	CurrentInputs    json.RawMessage `json:"current_inputs"`
	CurrentOutputs   json.RawMessage `json:"current_outputs"`
}

func (d *DataExecutionError) IsOOM() bool {
	switch d.ExceptionType {
	case ExceptionTypeOOM:
		return true
	case ExceptionTypeRE:
		if strings.Contains(d.ExceptionMessage, "out of memory") {
			return true
		}
	}
	return false
}
