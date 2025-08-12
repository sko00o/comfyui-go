package prompt

import (
	"math/rand"

	"github.com/sko00o/comfyui-go/node"
)

var _ Builder = (*A1111Base)(nil)
var _ Builder = (*A1111VAE)(nil)
var _ Builder = (*A1111Upscale)(nil)

type A1111Base struct {
	Base
	CLIPSkip int   `json:"clip_skip"`
	Loras    Loras `json:"loras"`

	EnableMetadata bool `json:"enable_metadata,omitempty"`
}

func (w *A1111Base) Build() Prompt {
	workflow, _, _ := w.BuildWithVAE("")
	return workflow
}

func (w *A1111Base) BuildWithVAE(vaeName string) (Prompt, node.PreNode, node.PreNode) {
	vae := node.PreNode{
		ID:   IDCheckpointLoader,
		Argc: 2,
	}
	var vaeNode node.Builder
	if vaeName != "" {
		vae = node.PreNode{
			ID:   IDVAELoader,
			Argc: 0,
		}
		vaeNode = node.VAELoader{
			VAEName: vaeName,
		}
	}

	if w.Seed < 0 {
		w.Seed = rand.Int()
	}
	stopAtCLIPLayer := -1
	if w.CLIPSkip > 0 {
		stopAtCLIPLayer = -w.CLIPSkip
	}

	workflow := Prompt{
		IDCheckpointLoader: node.CheckpointLoaderSimple{
			Checkpoint: w.Checkpoint,
		},
		IDCLIPSetLastLayer: node.CLIPSetLastLayer{
			StopAtCLIPLayer: stopAtCLIPLayer,
			CLIP: node.PreNode{
				ID:   IDCheckpointLoader,
				Argc: 1,
			},
		},
		IDEmptyLatentImage: node.EmptyLatentImage{
			Width:     w.ImageWidth,
			Height:    w.ImageHeight,
			BatchSize: w.BatchSize,
		},
		IDSaveImage: node.SaveImageWebsocket{
			EnableMetadata: w.EnableMetadata,
			Images: node.PreNode{
				ID:   IDVAEDecode,
				Argc: 0,
			},
		},
		IDVAEDecode: node.VAEDecode{
			Samples: node.PreNode{
				ID:   IDKSampler,
				Argc: 0,
			},
			VAE: vae,
		},
	}
	if vaeNode != nil {
		workflow[IDVAELoader] = vaeNode
	}

	loras, currModel, currCLIP := w.Loras.Build(IDLoraLoader, node.PreNode{
		ID:   IDCheckpointLoader,
		Argc: 0,
	}, node.PreNode{
		ID:   IDCLIPSetLastLayer,
		Argc: 0,
	})
	for id, node := range loras {
		workflow[id] = node
	}

	workflow[IDCLIPTextEncodePositive] = node.CLIPTextEncodeA1111{
		Text: w.PositivePrompt,
		CLIP: currCLIP,
	}
	workflow[IDCLIPTextEncodeNegative] = node.CLIPTextEncodeA1111{
		Text: w.NegativePrompt,
		CLIP: currCLIP,
	}
	workflow[IDKSampler] = node.KSamplerA1111{
		Seed:        w.Seed,
		Steps:       w.Steps,
		CFG:         w.CFG,
		SamplerName: w.SamplerName,
		Scheduler:   w.Scheduler,
		Denoise:     1,
		Model:       currModel,
		Positive: node.PreNode{
			ID:   IDCLIPTextEncodePositive,
			Argc: 0,
		},
		Negative: node.PreNode{
			ID:   IDCLIPTextEncodeNegative,
			Argc: 0,
		},
		LatentImage: node.PreNode{
			ID:   IDEmptyLatentImage,
			Argc: 0,
		},
	}

	return workflow, vae, currModel
}

type A1111VAE struct {
	A1111Base
	VAEName string `json:"vae_name"`
}

func (w *A1111VAE) Build() Prompt {
	workflow, _, _ := w.A1111Base.BuildWithVAE(w.VAEName)
	return workflow
}

type A1111Upscale struct {
	A1111Base
	VAEName string  `json:"vae_name"`
	Upscale Upscale `json:"upscale"`
}

func (w *A1111Upscale) Build() Prompt {
	workflow, vae, model := w.A1111Base.BuildWithVAE(w.VAEName)
	return w.Upscale.BuildOn(workflow, vae, model, w.ImageWidth, w.ImageHeight, w.Seed, w.CFG, w.EnableMetadata)
}
