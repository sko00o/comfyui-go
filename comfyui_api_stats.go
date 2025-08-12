package comfyui

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type StatsResp struct {
	System  SystemInfo   `json:"system"`
	Devices []DeviceInfo `json:"devices"`
}

type SystemInfo struct {
	OS             string   `json:"os"`
	RAMTotal       int      `json:"ram_total"`
	RAMFree        int      `json:"ram_free"`
	ComfyUIVersion string   `json:"comfyui_version"`
	PythonVersion  string   `json:"python_version"`
	PyTorchVersion string   `json:"pytorch_version"`
	EmbeddedPython bool     `json:"embedded_python"`
	Argv           []string `json:"argv"`
}

type DeviceInfo struct {
	Name           string `json:"name"`
	Type           string `json:"type"`
	Index          int    `json:"index"`
	VRAMTotal      int    `json:"vram_total"`
	VRAMFree       int    `json:"vram_free"`
	TorchVRAMTotal int    `json:"torch_vram_total"`
	TorchVRAMFree  int    `json:"torch_vram_free"`
}

func (c *Client) Stats() (*StatsResp, error) {
	var resp StatsResp
	if err := c.process(c.getJSON(ReqPathSystemStats, nil), func(p io.Reader, _ http.Header) error {
		if err := json.NewDecoder(p).Decode(&resp); err != nil {
			return fmt.Errorf("decode resp: %w", err)
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	return &resp, nil
}
