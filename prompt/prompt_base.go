package prompt

import (
	"math/rand"

	"github.com/sko00o/comfyui-go/node"
)

var _ Builder = (*Base)(nil)

type Base struct {
	Checkpoint     string  `json:"checkpoint"`
	PositivePrompt string  `json:"positive_prompt"`
	NegativePrompt string  `json:"negative_prompt"`
	ImageWidth     int     `json:"image_width"`
	ImageHeight    int     `json:"image_height"`
	BatchSize      int     `json:"batch_size"`
	Seed           int     `json:"seed"`
	Steps          int     `json:"steps"`
	CFG            float64 `json:"cfg"`
	SamplerName    string  `json:"sampler_name"`
	Scheduler      string  `json:"scheduler"`
	Type           string  `json:"type"`
}

func (w *Base) Build() Prompt {
	if w.Seed < 0 {
		w.Seed = rand.Int()
	}

	workflow := Prompt{
		IDKSampler: node.KSampler{
			Seed:        w.Seed,
			Steps:       w.Steps,
			CFG:         w.CFG,
			SamplerName: w.SamplerName,
			Scheduler:   w.Scheduler,
			Denoise:     1,
			Model: node.PreNode{
				ID:   IDCheckpointLoader,
				Argc: 0,
			},
			LatentImage: node.PreNode{
				ID:   IDEmptyLatentImage,
				Argc: 0,
			},
			Positive: node.PreNode{
				ID:   IDCLIPTextEncodePositive,
				Argc: 0,
			},
			Negative: node.PreNode{
				ID:   IDCLIPTextEncodeNegative,
				Argc: 0,
			},
		},
		IDCheckpointLoader: node.CheckpointLoaderSimple{
			Checkpoint: w.Checkpoint,
		},
		IDEmptyLatentImage: node.EmptyLatentImage{
			Width:     w.ImageWidth,
			Height:    w.ImageHeight,
			BatchSize: w.BatchSize,
		},
		IDCLIPTextEncodePositive: node.CLIPTextEncode{
			Text: w.PositivePrompt,
			CLIP: node.PreNode{
				ID:   IDCheckpointLoader,
				Argc: 1,
			},
		},
		IDCLIPTextEncodeNegative: node.CLIPTextEncode{
			Text: w.NegativePrompt,
			CLIP: node.PreNode{
				ID:   IDCheckpointLoader,
				Argc: 1,
			},
		},
		IDVAEDecode: node.VAEDecode{
			Samples: node.PreNode{
				ID:   IDKSampler,
				Argc: 0,
			},
			VAE: node.PreNode{
				ID:   IDCheckpointLoader,
				Argc: 2,
			},
		},
		IDSaveImage: node.PreviewImage{
			Images: node.PreNode{
				ID:   IDVAEDecode,
				Argc: 0,
			},
		},
	}
	return workflow
}

func (w Base) OutputNodeIDs() []string {
	return []string{IDSaveImage}
}
