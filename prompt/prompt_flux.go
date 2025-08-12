package prompt

import (
	"math/rand"
	"strings"

	"github.com/sko00o/comfyui-go/node"
)

var _ Builder = (*FluxBase)(nil)
var _ Builder = (*FluxUpscale)(nil)

type FluxBase struct {
	Base
	WeightDtype string `json:"weight_dtype"`
	Loras       Loras  `json:"loras"`
	VAEName     string `json:"vae_name"`

	EnableMetadata bool `json:"enable_metadata"`
}

func (w *FluxBase) Build() Prompt {
	workflow, _, _ := w.BuildWithVAE()
	return workflow
}

func (w *FluxBase) BuildWithVAE() (workflow Prompt, vae, model node.PreNode) {
	if w.VAEName == "" {
		w.VAEName = "ae.safetensors"
	}
	if w.WeightDtype == "" {
		w.WeightDtype = "fp8_e5m2"
	}
	if w.Seed < 0 {
		w.Seed = rand.Int()
	}

	guiderID := FluxIDPerpNegGuider
	if strings.TrimSpace(w.NegativePrompt) == "" {
		guiderID = FluxIDBasicGuider
	}

	model = node.PreNode{
		ID:   FluxIDModelSamplingFlux,
		Argc: 0,
	}

	workflow = Prompt{
		IDKSampler: node.GeneralNode{
			ClassType: "SamplerCustomAdvanced",
			Inputs: map[string]any{
				"noise": node.PreNode{
					ID:   FluxIDRandomNoise,
					Argc: 0,
				},
				"guider": node.PreNode{
					ID:   guiderID,
					Argc: 0,
				},
				"sampler": node.PreNode{
					ID:   FluxIDKSamplerSelect,
					Argc: 0,
				},
				"sigmas": node.PreNode{
					ID:   FluxIDBasicScheduler,
					Argc: 0,
				},
				"latent_image": node.PreNode{
					ID:   IDEmptyLatentImage,
					Argc: 0,
				},
			},
		},
		FluxIDRandomNoise: node.GeneralNode{
			ClassType: "RandomNoise",
			Inputs: map[string]any{
				"noise_seed": w.Seed,
			},
		},
		FluxIDBasicGuider: node.GeneralNode{
			ClassType: "BasicGuider",
			Inputs: map[string]any{
				"model": model,
				"conditioning": node.PreNode{
					ID:   FluxIDFluxPositiveGuidance,
					Argc: 0,
				},
			},
		},
		FluxIDPerpNegGuider: node.GeneralNode{
			ClassType: "PerpNegGuider",
			Inputs: map[string]any{
				"cfg":       1, // always 1 in Flux
				"neg_scale": 2,
				"model":     model,
				"positive": node.PreNode{
					ID:   FluxIDFluxPositiveGuidance,
					Argc: 0,
				},
				"negative": node.PreNode{
					ID:   FluxIDFluxNegativeGuidance,
					Argc: 0,
				},
				"empty_conditioning": node.PreNode{
					ID:   FluxIDEmptyCLIPTextEncode,
					Argc: 0,
				},
			},
		},
		FluxIDKSamplerSelect: node.GeneralNode{
			ClassType: "KSamplerSelect",
			Inputs: map[string]any{
				"sampler_name": w.SamplerName,
			},
		},
		FluxIDBasicScheduler: node.GeneralNode{
			ClassType: "BasicScheduler",
			Inputs: map[string]any{
				"scheduler": w.Scheduler,
				"steps":     w.Steps,
				"denoise":   1,
				"model":     model,
			},
		},
		IDEmptyLatentImage: node.GeneralNode{
			ClassType: "EmptyLatentImage",
			Inputs: map[string]any{
				"width":      w.ImageWidth,
				"height":     w.ImageHeight,
				"batch_size": w.BatchSize,
			},
		},

		FluxIDUNETLoader: node.GeneralNode{
			ClassType: "UNETLoader",
			Inputs: map[string]any{
				"unet_name":    w.Checkpoint,
				"weight_dtype": w.WeightDtype,
			},
		},
		FluxIDFluxPositiveGuidance: node.GeneralNode{
			ClassType: "FluxGuidance",
			Inputs: map[string]any{
				"guidance": w.CFG,
				"conditioning": node.PreNode{
					ID:   IDCLIPTextEncodePositive,
					Argc: 0,
				},
			},
		},
		FluxIDFluxNegativeGuidance: node.GeneralNode{
			ClassType: "FluxGuidance",
			Inputs: map[string]any{
				"guidance": w.CFG,
				"conditioning": node.PreNode{
					ID:   IDCLIPTextEncodeNegative,
					Argc: 0,
				},
			},
		},
		FluxIDDualCLIPLoader: node.GeneralNode{
			ClassType: "DualCLIPLoader",
			Inputs: map[string]any{
				"clip_name1": "t5xxl_fp16.safetensors",
				"clip_name2": "clip_l.safetensors",
				"type":       "flux",
			},
		},
		IDSaveImage: node.SaveImageWebsocket{
			Images: node.PreNode{
				ID:   IDVAEDecode,
				Argc: 0,
			},
			EnableMetadata: w.EnableMetadata,
		},
	}

	loras, currModel, currCLIP := w.Loras.Build(IDLoraLoader, node.PreNode{
		ID:   FluxIDUNETLoader,
		Argc: 0,
	}, node.PreNode{
		ID:   FluxIDDualCLIPLoader,
		Argc: 0,
	})
	for id, node := range loras {
		workflow[id] = node
	}

	workflow[FluxIDModelSamplingFlux] = node.GeneralNode{
		ClassType: "ModelSamplingFlux",
		Inputs: map[string]any{
			"max_shift":  1.15,
			"base_shift": 0.5,
			"width":      w.ImageWidth,
			"height":     w.ImageHeight,
			"model":      currModel,
		},
	}
	workflow[FluxIDEmptyCLIPTextEncode] = node.GeneralNode{
		ClassType: "CLIPTextEncode",
		Inputs: map[string]any{
			"text": "",
			"clip": currCLIP,
		},
	}
	workflow[IDCLIPTextEncodePositive] = node.GeneralNode{
		ClassType: "CLIPTextEncode",
		Inputs: map[string]any{
			"text": w.PositivePrompt,
			"clip": currCLIP,
		},
	}
	workflow[IDCLIPTextEncodeNegative] = node.GeneralNode{
		ClassType: "CLIPTextEncode",
		Inputs: map[string]any{
			"text": w.NegativePrompt,
			"clip": currCLIP,
		},
	}

	vae = node.PreNode{
		ID:   IDVAELoader,
		Argc: 0,
	}
	workflow[IDVAELoader] = node.GeneralNode{
		ClassType: "VAELoader",
		Inputs: map[string]any{
			"vae_name": w.VAEName,
		},
	}
	workflow[IDVAEDecode] = node.GeneralNode{
		ClassType: "VAEDecode",
		Inputs: map[string]any{
			"samples": node.PreNode{
				ID:   IDKSampler,
				Argc: 0,
			},
			"vae": vae,
		},
	}

	return workflow, vae, model
}

func (w *FluxBase) OutputNodeIDs() []string {
	return []string{IDSaveImage}
}

type FluxUpscale struct {
	FluxBase
	Upscale Upscale `json:"upscale"`
}

func (w *FluxUpscale) Build() Prompt {
	workflow, vae, model := w.FluxBase.BuildWithVAE()
	return w.Upscale.BuildOn(workflow, vae, model, w.ImageWidth, w.ImageHeight, w.Seed, 1, w.EnableMetadata)
}
