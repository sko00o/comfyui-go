package prompt

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPromptBuilder(t *testing.T) {
	const baseDir = "./test"

	tests := []struct {
		workflow   Builder
		expectFile string
	}{
		{
			workflow: &Base{
				Checkpoint:     "v1-5-pruned-emaonly.safetensors",
				PositivePrompt: "beautiful scenery nature glass bottle landscape, purple galaxy bottle,",
				NegativePrompt: "text, watermark",
				ImageWidth:     512,
				ImageHeight:    512,
				BatchSize:      1,
				Seed:           1034428433385519,
				Steps:          20,
				CFG:            8,
				SamplerName:    "euler",
				Scheduler:      "normal",
			},
			expectFile: "base_workflow_api.json",
		},
		{
			workflow: &A1111Base{
				Base: Base{
					Checkpoint:     "toonyou_beta3.safetensors",
					PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
					NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
					ImageWidth:     512,
					ImageHeight:    768,
					BatchSize:      1,
					Seed:           1099633726,
					Steps:          20,
					CFG:            8,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "karras",
				},
				CLIPSkip: 2,
				Loras: []Lora{
					{
						LoraName:      "alice_liddell_v2.safetensors",
						StrengthModel: 1,
						StrengthCLIP:  1,
					},
				},
			},
			expectFile: "a1111_base_workflow_api.json",
		},
		{
			workflow: &A1111Base{
				Base: Base{
					Checkpoint:     "toonyou_beta3.safetensors",
					PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
					NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
					ImageWidth:     512,
					ImageHeight:    768,
					BatchSize:      1,
					Seed:           1099633726,
					Steps:          20,
					CFG:            8,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "karras",
				},
				CLIPSkip: 2,
				Loras:    []Lora{},
			},
			expectFile: "a1111_no_lora_workflow_api.json",
		},
		{
			workflow: &A1111Base{
				EnableMetadata: true,
				Base: Base{
					Checkpoint:     "toonyou_beta3.safetensors",
					PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
					NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
					ImageWidth:     512,
					ImageHeight:    768,
					BatchSize:      1,
					Seed:           1099633726,
					Steps:          20,
					CFG:            8,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "karras",
				},
				CLIPSkip: 2,
				Loras: []Lora{
					{
						LoraName:      "alice_liddell_v2.safetensors",
						StrengthModel: 1,
						StrengthCLIP:  1,
					},
				},
			},
			expectFile: "a1111_base_workflow_metadata_api.json",
		},
		{
			workflow: &A1111VAE{
				A1111Base: A1111Base{
					Base: Base{
						Checkpoint:     "toonyou_beta3.safetensors",
						PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
						NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
						ImageWidth:     512,
						ImageHeight:    768,
						BatchSize:      1,
						Seed:           1099633726,
						Steps:          20,
						CFG:            8,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "karras",
					},
					CLIPSkip: 2,
					Loras: []Lora{
						{
							LoraName:      "alice_liddell_v2.safetensors",
							StrengthModel: 1,
							StrengthCLIP:  1,
						},
					},
				},
				VAEName: "animevae.pt",
			},
			expectFile: "a1111_vae_workflow_api.json",
		},
		{
			workflow: &A1111Upscale{
				A1111Base: A1111Base{
					Base: Base{
						Checkpoint:     "toonyou_beta3.safetensors",
						PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
						NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
						ImageWidth:     512,
						ImageHeight:    768,
						BatchSize:      1,
						Seed:           1099633726,
						Steps:          20,
						CFG:            8,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "karras",
					},
					CLIPSkip: 2,
					Loras: []Lora{
						{
							LoraName:      "alice_liddell_v2.safetensors",
							StrengthModel: 1,
							StrengthCLIP:  1,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 2,
					Model:   "4x-AnimeSharp.pth",
					Steps:   20,
					Denoise: 0.5,
				},
			},
			expectFile: "a1111_upscale_workflow_api.json",
		},
		{
			workflow: &A1111Upscale{
				A1111Base: A1111Base{
					Base: Base{
						Checkpoint:     "toonyou_beta3.safetensors",
						PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
						NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
						ImageWidth:     512,
						ImageHeight:    768,
						BatchSize:      1,
						Seed:           1099633726,
						Steps:          20,
						CFG:            8,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "karras",
					},
					CLIPSkip: 2,
					Loras: []Lora{
						{
							LoraName:      "alice_liddell_v2.safetensors",
							StrengthModel: 1,
							StrengthCLIP:  1,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 2,
					Model:   "nearest-exact",
					Steps:   20,
					Denoise: 0.5,
				},
			},
			expectFile: "a1111_upscale_method_workflow_api.json",
		},
		{
			workflow: &A1111Upscale{
				A1111Base: A1111Base{
					Base: Base{
						Checkpoint:     "toonyou_beta3.safetensors",
						PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
						NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
						ImageWidth:     512,
						ImageHeight:    768,
						BatchSize:      1,
						Seed:           1099633726,
						Steps:          20,
						CFG:            8,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "karras",
					},
					CLIPSkip: 2,
					Loras: []Lora{
						{
							LoraName:      "alice_liddell_v2.safetensors",
							StrengthModel: 1,
							StrengthCLIP:  1,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 2,
					Model:   "latent-nearest-exact",
					Steps:   20,
					Denoise: 0.5,
				},
			},
			expectFile: "a1111_upscale_latent_workflow_api.json",
		},
		{
			workflow: &A1111Upscale{
				A1111Base: A1111Base{
					Base: Base{
						Checkpoint:     "toonyou_beta3.safetensors",
						PositivePrompt: "(masterpiece:1.4),(best qualit:1.4),(high resolution:1.4),alice liddell,blue dress,white apron,black hairband,smile,looking at viewer",
						NegativePrompt: "nsfw, (bad anatomy, worst quality, low quality:1.4), watermark, signature, username, patreon, monochrome, zombie, squinting, badhandv4, easynegative",
						ImageWidth:     512,
						ImageHeight:    768,
						BatchSize:      1,
						Seed:           1099633726,
						Steps:          20,
						CFG:            8,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "karras",
					},
					CLIPSkip: 2,
					Loras: []Lora{
						{
							LoraName:      "alice_liddell_v2.safetensors",
							StrengthModel: 1,
							StrengthCLIP:  1,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 2,
					Model:   "4x-AnimeSharp.pth",
					Steps:   20,
					Denoise: 0.5,
				},
				VAEName: "animevae.pt",
			},
			expectFile: "a1111_vae_upscale_workflow_api.json",
		},
		{
			workflow: &FluxBase{
				Base: Base{
					Checkpoint:     "flux1-dev.safetensors",
					PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
					NegativePrompt: "cgi, doll, deformed, disfigured, poorly drawn, bad anatomy, red background",
					ImageWidth:     832,
					ImageHeight:    1248,
					BatchSize:      1,
					Seed:           634892015342167,
					Steps:          30,
					CFG:            3.5,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				WeightDtype: "default",
				VAEName:     "ae.safetensors",
				Loras: []Lora{
					{
						LoraName:      "Flux/aesthetic2-cdo-0.5.safetensors",
						StrengthModel: 0.8,
						StrengthCLIP:  0.8,
					},
					{
						LoraName:      "Flux/flux_pam_crystal.safetensors",
						StrengthModel: 0.4,
						StrengthCLIP:  0.4,
					},
				},
			},
			expectFile: "flux_base_workflow_api.json",
		},
		{
			workflow: &FluxBase{
				Base: Base{
					Checkpoint:     "flux1-dev-fp8-e5m2.safetensors",
					PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
					NegativePrompt: "cgi, doll, deformed, disfigured, poorly drawn, bad anatomy, red background",
					ImageWidth:     832,
					ImageHeight:    1248,
					BatchSize:      1,
					Seed:           634892015342167,
					Steps:          30,
					CFG:            3.5,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				WeightDtype: "",
				VAEName:     "ae.safetensors",
				Loras: []Lora{
					{
						LoraName:      "Flux/aesthetic2-cdo-0.5.safetensors",
						StrengthModel: 0.8,
						StrengthCLIP:  0.8,
					},
					{
						LoraName:      "Flux/flux_pam_crystal.safetensors",
						StrengthModel: 0.4,
						StrengthCLIP:  0.4,
					},
				},
			},
			expectFile: "flux_base_fp8_workflow_api.json",
		},
		{
			workflow: &FluxBase{
				Base: Base{
					Checkpoint:     "flux1-dev.safetensors",
					PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
					ImageWidth:     832,
					ImageHeight:    1248,
					BatchSize:      1,
					Seed:           634892015342167,
					Steps:          30,
					CFG:            3.5,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				WeightDtype: "default",
				VAEName:     "ae.safetensors",
				Loras: []Lora{
					{
						LoraName:      "Flux/flux_pam_crystal.safetensors",
						StrengthModel: 0.4,
						StrengthCLIP:  0.4,
					},
				},
			},
			expectFile: "flux_no_neg_workflow_api.json",
		},
		{
			workflow: &FluxBase{
				Base: Base{
					Checkpoint:     "flux1-dev.safetensors",
					PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
					ImageWidth:     832,
					ImageHeight:    1248,
					BatchSize:      1,
					Seed:           634892015342167,
					Steps:          30,
					CFG:            3.5,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				WeightDtype: "default",
				VAEName:     "ae.safetensors",
				Loras:       []Lora{},
			},
			expectFile: "flux_no_lora_workflow_api.json",
		},
		{
			workflow: &FluxUpscale{
				FluxBase: FluxBase{
					Base: Base{
						Checkpoint:     "flux1-dev.safetensors",
						PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
						NegativePrompt: "cgi, doll, deformed, disfigured, poorly drawn, bad anatomy, red background",
						ImageWidth:     832,
						ImageHeight:    1248,
						BatchSize:      1,
						Seed:           634892015342167,
						Steps:          30,
						CFG:            3.5,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "sgm_uniform",
					},
					WeightDtype: "default",
					VAEName:     "ae.safetensors",
					Loras: []Lora{
						{
							LoraName:      "Flux/aesthetic2-cdo-0.5.safetensors",
							StrengthModel: 0.8,
							StrengthCLIP:  0.8,
						},
						{
							LoraName:      "Flux/flux_pam_crystal.safetensors",
							StrengthModel: 0.4,
							StrengthCLIP:  0.4,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 1,
					Model:   "latent-nearest-exact",
					Steps:   20,
					Denoise: 0.5,
				},
			},
			expectFile: "flux_upscale_latent_workflow_api.json",
		},
		{
			workflow: &FluxUpscale{
				FluxBase: FluxBase{
					Base: Base{
						Checkpoint:     "flux1-dev.safetensors",
						PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
						NegativePrompt: "cgi, doll, deformed, disfigured, poorly drawn, bad anatomy, red background",
						ImageWidth:     832,
						ImageHeight:    1248,
						BatchSize:      1,
						Seed:           634892015342167,
						Steps:          30,
						CFG:            3.5,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "sgm_uniform",
					},
					WeightDtype: "default",
					VAEName:     "ae.safetensors",
					Loras: []Lora{
						{
							LoraName:      "Flux/aesthetic2-cdo-0.5.safetensors",
							StrengthModel: 0.8,
							StrengthCLIP:  0.8,
						},
						{
							LoraName:      "Flux/flux_pam_crystal.safetensors",
							StrengthModel: 0.4,
							StrengthCLIP:  0.4,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 1,
					Model:   "nearest-exact",
					Steps:   20,
					Denoise: 0.5,
				},
			},
			expectFile: "flux_upscale_method_workflow_api.json",
		},
		{
			workflow: &FluxUpscale{
				FluxBase: FluxBase{
					Base: Base{
						Checkpoint:     "flux1-dev.safetensors",
						PositivePrompt: "Portrait of a beautiful woman, 22 years old woman in a red latex outfit, dynamic lights, text \"Good\" made of crystal on the wall",
						NegativePrompt: "cgi, doll, deformed, disfigured, poorly drawn, bad anatomy, red background",
						ImageWidth:     832,
						ImageHeight:    1248,
						BatchSize:      1,
						Seed:           634892015342167,
						Steps:          30,
						CFG:            3.5,
						SamplerName:    "dpmpp_2m",
						Scheduler:      "sgm_uniform",
					},
					WeightDtype: "default",
					VAEName:     "ae.safetensors",
					Loras: []Lora{
						{
							LoraName:      "Flux/aesthetic2-cdo-0.5.safetensors",
							StrengthModel: 0.8,
							StrengthCLIP:  0.8,
						},
						{
							LoraName:      "Flux/flux_pam_crystal.safetensors",
							StrengthModel: 0.4,
							StrengthCLIP:  0.4,
						},
					},
				},
				Upscale: Upscale{
					ScaleBy: 1,
					Model:   "4x-AnimeSharp.pth",
					Steps:   20,
					Denoise: 0.5,
				},
			},
			expectFile: "flux_upscale_model_workflow_api.json",
		},
		{
			workflow: &SD3Base{
				Base: Base{
					Checkpoint:     "sd3.5_large_turbo.safetensors",
					PositivePrompt: "beautiful scenery nature glass bottle landscape, purple galaxy bottle,",
					NegativePrompt: "",
					ImageWidth:     1024,
					ImageHeight:    1024,
					BatchSize:      1,
					Seed:           1042540244833452,
					Steps:          4,
					CFG:            1,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				Loras: []Lora{
					{
						LoraName:      "SD35-lora-Futuristic-Bzonze-Colored.safetensors",
						StrengthModel: 1,
						StrengthCLIP:  1,
					},
				},
			},
			expectFile: "sd3.5_base_workflow_api.json",
		},
		{
			workflow: &SD3Base{
				Base: Base{
					Checkpoint:     "sd3.5_large_turbo.safetensors",
					PositivePrompt: "beautiful scenery nature glass bottle landscape, purple galaxy bottle,",
					NegativePrompt: "",
					ImageWidth:     1024,
					ImageHeight:    1024,
					BatchSize:      1,
					Seed:           1042540244833452,
					Steps:          4,
					CFG:            1,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				Loras: []Lora{},
			},
			expectFile: "sd3.5_no_lora_workflow_api.json",
		},
		{
			workflow: &SD3Base{
				Base: Base{
					Checkpoint:     "sd3.5_large_turbo.safetensors",
					PositivePrompt: "beautiful scenery nature glass bottle landscape, purple galaxy bottle,",
					NegativePrompt: "",
					ImageWidth:     1024,
					ImageHeight:    1024,
					BatchSize:      1,
					Seed:           1042540244833452,
					Steps:          4,
					CFG:            1,
					SamplerName:    "dpmpp_2m",
					Scheduler:      "sgm_uniform",
				},
				VAEName: "ae.safetensors",
				Loras:   []Lora{},
			},
			expectFile: "sd3.5_vae_workflow_api.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expectFile, func(t *testing.T) {
			wantB, err := os.ReadFile(filepath.Join(baseDir, tt.expectFile))
			assert.NoError(t, err)
			w := tt.workflow.Build()
			gotB, err := json.Marshal(w)
			assert.NoError(t, err)
			assert.JSONEq(t, string(wantB), string(gotB))
		})
	}
}
