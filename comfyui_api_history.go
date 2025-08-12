package comfyui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/sko00o/comfyui-go/ws/message"
)

func (c *Client) GetHistory(maxItems int) (HistoryResp, error) {
	params := url.Values{}
	if maxItems > 0 {
		params.Add("max_items", strconv.Itoa(maxItems))
	}
	var resp HistoryResp
	if err := c.process(c.getJSON(ReqPathHistory, params), func(p io.Reader, _ http.Header) error {
		if err := json.NewDecoder(p).Decode(&resp); err != nil {
			return fmt.Errorf("decode resp: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	return resp, nil
}

// HistoryResp has prompt_id in keys
type HistoryResp map[string]HistoryObj

type HistoryObj struct {
	// Outputs has node in keys
	Outputs message.MapOutput `json:"outputs"`

	Prompt PromptObj `json:"prompt"`
	Status StatusObj `json:"status"`
}

type PromptObj struct {
	Num           uint64
	PromptID      string
	Workflow      json.RawMessage
	ExtraData     json.RawMessage
	OutputNodeIDs []string
}

func (o *PromptObj) UnmarshalJSON(p []byte) error {
	var temp []json.RawMessage
	if err := json.Unmarshal(p, &temp); err != nil {
		return err
	}
	if len(temp) != 5 {
		return fmt.Errorf("unexpected array length")
	}

	for _, v := range []struct {
		data   json.RawMessage
		target any
	}{
		{temp[0], &o.Num},
		{temp[1], &o.PromptID},
		{temp[2], &o.Workflow},
		{temp[3], &o.ExtraData},
		{temp[4], &o.OutputNodeIDs},
	} {
		if err := json.Unmarshal(v.data, v.target); err != nil {
			return err
		}
	}
	return nil
}

type StatusObj struct {
	Completed bool         `json:"completed"`
	StatusStr string       `json:"status_str"`
	Messages  []MessageObj `json:"messages"`
}

type MessageObj message.Message

func (o *MessageObj) UnmarshalJSON(p []byte) error {
	var temp []json.RawMessage
	if err := json.Unmarshal(p, &temp); err != nil {
		return err
	}
	if len(temp) != 2 {
		return fmt.Errorf("unexpected array length")
	}

	o.Data = &message.DataExecution{}
	for _, v := range []struct {
		data   json.RawMessage
		target any
	}{
		{temp[0], &o.Type},
		{temp[1], &o.Data},
	} {
		if err := json.Unmarshal(v.data, v.target); err != nil {
			return err
		}
	}
	return nil
}
