package comfyui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sko00o/comfyui-go/ws/message"
)

type ReqPath string

const (
	// Core API
	// ref: https://docs.comfy.org/development/comfyui-server/comms_routes
	ReqPathPrompt      ReqPath = "/api/prompt"
	ReqPathHistory     ReqPath = "/api/history"
	ReqPathView        ReqPath = "/api/view"
	ReqPathSystemStats ReqPath = "/api/system_stats"
	//ReqPathViewMetadata ReqPath = "/view_metadata"
	//ReqPathEmbeddings   ReqPath = "/embeddings"
	//ReqPathExtensions   ReqPath = "/extensions"
	//ReqPathInterrupt    ReqPath = "/interrupt"
	//ReqPathQueue        ReqPath = "/queue"
	//ReqPathObjectInfo   ReqPath = "/object_info"
	//ReqPathUploadImage  ReqPath = "/upload/image"
	//ReqPathUploadMask   ReqPath = "/upload/mask"
	//ReqPathQueue        ReqPath = "/queue"
	//ReqPathInterrupt    ReqPath = "/interrupt"
	//ReqPathFree         ReqPath = "/free"

	// API in VHS
	ReqPathViewVideo ReqPath = "/api/vhs/viewvideo"

	// API in ComfyUI-Manager
	ReqPathReboot ReqPath = "/api/manager/reboot"
)

type QueuePromptResp struct {
	PromptID   string          `json:"prompt_id"`
	Number     int             `json:"number"`
	NodeErrors json.RawMessage `json:"node_errors,omitempty"`
}

// Prompt submit a prompt to the queue
func (c *Client) Prompt(data map[string]any) (*QueuePromptResp, error) {
	var respContent QueuePromptResp
	if err := c.process(c.postJSON(ReqPathPrompt, data), func(p io.Reader, _ http.Header) error {
		if err := json.NewDecoder(p).Decode(&respContent); err != nil {
			return fmt.Errorf("decode resp: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	return &respContent, nil
}

type GetPromptResp struct {
	ExecInfo ExecInfo `json:"exec_info"`
}

type ExecInfo struct {
	QueueRemaining int `json:"queue_remaining"`
}

// GetPrompt retrieve current status
func (c *Client) GetPrompt() (*GetPromptResp, error) {
	var resp GetPromptResp
	if err := c.process(c.getJSON(ReqPathPrompt, nil), func(p io.Reader, _ http.Header) error {
		if err := json.NewDecoder(p).Decode(&resp); err != nil {
			return fmt.Errorf("decode resp: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	return &resp, nil
}

func (c *Client) GetView(f message.FileInfo, handle handleRespFunc) error {
	params := url.Values{}
	params.Add("filename", f.Filename)
	params.Add("subfolder", f.Subfolder)
	params.Add("type", f.Type)
	return c.process(c.getJSON(ReqPathView, params), handle)
}

// always return .webm in this response
func (c *Client) GetViewVideo(f message.FileInfo, handle handleRespFunc) error {
	params := url.Values{}
	params.Add("filename", f.Filename)
	params.Add("subfolder", f.Subfolder)
	params.Add("type", f.Type)
	return c.process(c.getJSON(ReqPathViewVideo, params), handle)
}

func (c *Client) Reboot() error {
	return c.process(c.getJSON(ReqPathReboot, nil), nil)
}
