package prompt

import (
	"strings"

	"github.com/sko00o/comfyui-go/node"
)

type Upscale struct {
	Model       string  `json:"model"`
	Steps       int     `json:"steps"`
	Denoise     float64 `json:"denoise"`
	ScaleBy     float64 `json:"scale_by"`
	ImageWidth  int     `json:"image_width"`
	ImageHeight int     `json:"image_height"`
}

const latentPrefix = "latent-"

const (
	MethodNearestExact = "nearest-exact"
	MethodBilinear     = "bilinear"
	MethodArea         = "area"
	MethodBicubic      = "bicubic"

	// ImageUpscale only
	MethodLanczos = "lanczos"

	// LatentUpscale only
	MethodBislerp = "bislerp"
)

func isUpscaleMethod(s string) bool {
	switch s {
	// Tips:
	//   antialiased -> bilinear
	//   nearest -> nearest-exact
	case MethodNearestExact, MethodBilinear, MethodArea, MethodBicubic, MethodLanczos:
		return true
	}
	return false
}

func isLatentUpscaleMethod(s string) bool {
	switch s {
	case MethodNearestExact, MethodBilinear, MethodArea, MethodBicubic, MethodBislerp:
		return true
	}
	return false
}

func (w *Upscale) BuildOn(workflow Prompt, vae, model node.PreNode, srcImageWidth, srcImageHeight int, seed int, cfg float64, enableMetadata bool) Prompt {
	// FIX: default denoise
	if w.Denoise == 0 {
		w.Denoise = 0.5
	}

	newWidth := w.ImageWidth
	newHeight := w.ImageHeight
	if w.ScaleBy > 0 {
		newWidth = int(float64(srcImageWidth) * w.ScaleBy)
		newHeight = int(float64(srcImageHeight) * w.ScaleBy)
	}

	upscaleMethod := w.Model

	var latentFrom node.PreNode
	if m := strings.TrimPrefix(upscaleMethod, latentPrefix); m != upscaleMethod && isLatentUpscaleMethod(m) {
		// Using LatentUpscale
		workflow[IDLatentScale] = node.LatentUpscale{
			UpscaleMethod: m,
			Width:         newWidth,
			Height:        newHeight,
			Crop:          "disabled",
			Samples: node.PreNode{
				ID:   IDKSampler,
				Argc: 0,
			},
		}
		latentFrom = node.PreNode{
			ID:   IDLatentScale,
			Argc: 0,
		}
	} else {
		// Using ImageUpscale
		var imageFrom node.PreNode
		if isUpscaleMethod(upscaleMethod) {
			imageFrom = node.PreNode{
				ID:   IDVAEDecode,
				Argc: 0,
			}
		} else {
			workflow[IDImageUpscaleWithModel] = node.ImageUpscaleWithModel{
				UpscaleModel: node.PreNode{
					ID:   IDUpscaleModelLoader,
					Argc: 0,
				},
				Image: node.PreNode{
					ID:   IDVAEDecode,
					Argc: 0,
				},
			}
			workflow[IDUpscaleModelLoader] = node.UpscaleModelLoader{
				ModelName: w.Model,
			}
			imageFrom = node.PreNode{
				ID:   IDImageUpscaleWithModel,
				Argc: 0,
			}
			upscaleMethod = MethodNearestExact
		}
		workflow[IDImageScale] = node.ImageScale{
			UpscaleMethod: upscaleMethod,
			Width:         newWidth,
			Height:        newHeight,
			Crop:          "disabled",
			Image:         imageFrom,
		}
		workflow[IDVAEEncode] = node.VAEEncode{
			Pixels: node.PreNode{
				ID:   IDImageScale,
				Argc: 0,
			},
			VAE: vae,
		}
		latentFrom = node.PreNode{
			ID:   IDVAEEncode,
			Argc: 0,
		}
	}

	workflow[IDKSamplerUpscale] = node.KSampler{
		Seed:        seed,
		Steps:       w.Steps,
		CFG:         cfg,
		SamplerName: "euler",
		Scheduler:   "normal",
		Denoise:     w.Denoise,
		Model:       model,
		Positive: node.PreNode{
			ID:   IDCLIPTextEncodePositive,
			Argc: 0,
		},
		Negative: node.PreNode{
			ID:   IDCLIPTextEncodeNegative,
			Argc: 0,
		},
		LatentImage: latentFrom,
	}
	workflow[IDVAEDecodeUpscale] = node.VAEDecode{
		Samples: node.PreNode{
			ID:   IDKSamplerUpscale,
			Argc: 0,
		},
		VAE: vae,
	}
	workflow[IDSaveImage] = node.SaveImageWebsocket{
		EnableMetadata: enableMetadata,
		Images: node.PreNode{
			ID:   IDVAEDecodeUpscale,
			Argc: 0,
		},
	}
	return workflow
}
