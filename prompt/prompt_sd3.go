package prompt

import (
	"math/rand"

	"github.com/sko00o/comfyui-go/node"
)

var _ Builder = (*SD3Base)(nil)
var _ Builder = (*SD3Upscale)(nil)

type SD3Base struct {
	Base
	Loras   Loras  `json:"loras"`
	VAEName string `json:"vae_name"`

	EnableMetadata bool `json:"enable_metadata"`
}

func (w *SD3Base) BuildWithVAE() (workflow Prompt, vae, model node.PreNode) {
	if w.Seed < 0 {
		w.Seed = rand.Int()
	}

	workflow = Prompt{
		IDCheckpointLoader: node.GeneralNode{
			ClassType: "CheckpointLoaderSimple",
			Inputs: map[string]any{
				"ckpt_name": w.Checkpoint,
			},
		},

		IDTripleCLIPLoader: node.GeneralNode{
			ClassType: "TripleCLIPLoader",
			Inputs: map[string]any{
				"clip_name1": "clip_g.safetensors",
				"clip_name2": "clip_l.safetensors",
				"clip_name3": "t5xxl_fp16.safetensors",
			},
		},
		IDModelSamplingSD3: node.GeneralNode{
			ClassType: "ModelSamplingSD3",
			Inputs: map[string]any{
				"shift": 3,
				"model": node.PreNode{
					ID:   IDCheckpointLoader,
					Argc: 0,
				},
			},
		},
		IDSaveImage: node.SaveImageWebsocket{
			Images: node.PreNode{
				ID:   IDVAEDecode,
				Argc: 0,
			},
			EnableMetadata: w.EnableMetadata,
		},
		IDConditioningZeroOut: node.GeneralNode{
			ClassType: "ConditioningZeroOut",
			Inputs: map[string]any{
				"conditioning": node.PreNode{
					ID:   IDCLIPTextEncodeNegative,
					Argc: 0,
				},
			},
		},
		IDConditioningSetTimestepRange1: node.GeneralNode{
			ClassType: "ConditioningSetTimestepRange",
			Inputs: map[string]any{
				"start": 0.1,
				"end":   1,
				"conditioning": node.PreNode{
					ID:   IDConditioningZeroOut,
					Argc: 0,
				},
			},
		},
		IDConditioningCombine: node.GeneralNode{
			ClassType: "ConditioningCombine",
			Inputs: map[string]any{
				"conditioning_1": node.PreNode{
					ID:   IDConditioningSetTimestepRange1,
					Argc: 0,
				},
				"conditioning_2": node.PreNode{
					ID:   IDConditioningSetTimestepRange2,
					Argc: 0,
				},
			},
		},
		IDConditioningSetTimestepRange2: node.GeneralNode{
			ClassType: "ConditioningSetTimestepRange",
			Inputs: map[string]any{
				"start": 0,
				"end":   0.1,
				"conditioning": node.PreNode{
					ID:   IDCLIPTextEncodeNegative,
					Argc: 0,
				},
			},
		},
		IDEmptySD3LatentImage: node.GeneralNode{
			ClassType: "EmptySD3LatentImage",
			Inputs: map[string]any{
				"width":      w.ImageWidth,
				"height":     w.ImageHeight,
				"batch_size": w.BatchSize,
			},
		},
	}

	vae = node.PreNode{
		ID:   IDCheckpointLoader,
		Argc: 2,
	}
	if w.VAEName != "" {
		workflow[IDVAELoader] = node.GeneralNode{
			ClassType: "VAELoader",
			Inputs: map[string]any{
				"vae_name": w.VAEName,
			},
		}
		vae = node.PreNode{
			ID:   IDVAELoader,
			Argc: 0,
		}
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

	loras, currModel, currCLIP := w.Loras.Build(IDLoraLoader, node.PreNode{
		ID:   IDModelSamplingSD3,
		Argc: 0,
	}, node.PreNode{
		ID:   IDTripleCLIPLoader,
		Argc: 0,
	})
	for id, node := range loras {
		workflow[id] = node
	}

	workflow[IDKSampler] = node.GeneralNode{
		ClassType: "KSampler",
		Inputs: map[string]any{
			"seed":         w.Seed,
			"steps":        w.Steps,
			"cfg":          w.CFG,
			"sampler_name": w.SamplerName,
			"scheduler":    w.Scheduler,
			"denoise":      1,
			"model":        currModel,
			"positive": node.PreNode{
				ID:   IDCLIPTextEncodePositive,
				Argc: 0,
			},
			"negative": node.PreNode{
				ID:   IDConditioningCombine,
				Argc: 0,
			},
			"latent_image": node.PreNode{
				ID:   IDEmptySD3LatentImage,
				Argc: 0,
			},
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
	return workflow, vae, currModel
}

func (w *SD3Base) Build() Prompt {
	workflow, _, _ := w.BuildWithVAE()
	return workflow
}

type SD3Upscale struct {
	SD3Base
	Upscale Upscale `json:"upscale"`
}

func (w *SD3Upscale) Build() Prompt {
	workflow, vae, model := w.SD3Base.BuildWithVAE()
	return w.Upscale.BuildOn(workflow, vae, model, w.ImageWidth, w.ImageHeight, w.Seed, w.CFG, w.EnableMetadata)
}
