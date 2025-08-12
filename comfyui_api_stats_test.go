package comfyui

import (
	"encoding/json"
	"reflect"
	"testing"
)

// test unmarshal StatsResp
func TestUnmarshalStatsResp(t *testing.T) {
	jsonStr := `
{
  "system": {
    "os": "posix",
    "ram_total": 211062931456,
    "ram_free": 200273879040,
    "comfyui_version": "0.3.14",
    "python_version": "3.11.11 | packaged by conda-forge | (main, Dec  5 2024, 14:17:24) [GCC 13.3.0]",
    "pytorch_version": "2.6.0+cu124",
    "embedded_python": false,
    "argv": [
      "main.py",
      "--listen",
      "--port",
      "8188",
      "--preview-method",
      "auto",
      "--use-pytorch-cross-attention",
      "--preview-method",
      "none"
    ]
  },
  "devices": [
    {
      "name": "cuda:0 NVIDIA GeForce RTX 4090 : cudaMallocAsync",
      "type": "cuda",
      "index": 0,
      "vram_total": 25282281472,
      "vram_free": 24833294336,
      "torch_vram_total": 0,
      "torch_vram_free": 0
    }
  ]
}`

	want := StatsResp{
		System: SystemInfo{
			OS:             "posix",
			RAMTotal:       211062931456,
			RAMFree:        200273879040,
			ComfyUIVersion: "0.3.14",
			PythonVersion:  "3.11.11 | packaged by conda-forge | (main, Dec  5 2024, 14:17:24) [GCC 13.3.0]",
			PyTorchVersion: "2.6.0+cu124",
			EmbeddedPython: false,
			Argv: []string{
				"main.py",
				"--listen",
				"--port",
				"8188",
				"--preview-method",
				"auto",
				"--use-pytorch-cross-attention",
				"--preview-method",
				"none",
			},
		},
		Devices: []DeviceInfo{
			{
				Name:           "cuda:0 NVIDIA GeForce RTX 4090 : cudaMallocAsync",
				Type:           "cuda",
				Index:          0,
				VRAMTotal:      25282281472,
				VRAMFree:       24833294336,
				TorchVRAMTotal: 0,
				TorchVRAMFree:  0,
			},
		},
	}

	var got StatsResp
	err := json.Unmarshal([]byte(jsonStr), &got)
	if err != nil {
		t.Fatalf("failed to unmarshal stats resp: %v", err)
	}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got: %+v, want: %+v", got, want)
	}
}
