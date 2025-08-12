package node

var _ Builder = (*KSampler)(nil)
var _ Builder = (*CLIPTextEncode)(nil)
var _ Builder = (*PreviewImage)(nil)
var _ Builder = (*CLIPSetLastLayer)(nil)
var _ Builder = (*ImageScale)(nil)
var _ Builder = (*LatentUpscale)(nil)
var _ Builder = (*ImageScaleBy)(nil)
var _ Builder = (*VAEEncode)(nil)
var _ Builder = (*ImageUpscaleWithModel)(nil)
var _ Builder = (*UpscaleModelLoader)(nil)
var _ Builder = (*CheckpointLoaderSimple)(nil)
var _ Builder = (*EmptyLatentImage)(nil)
var _ Builder = (*VAELoader)(nil)
var _ Builder = (*VAEDecode)(nil)

type KSampler struct {
	Seed        int     `json:"seed"`
	Steps       int     `json:"steps"`
	CFG         float64 `json:"cfg"`
	SamplerName string  `json:"sampler_name"`
	Scheduler   string  `json:"scheduler"`
	Denoise     float64 `json:"denoise"`
	Model       PreNode `json:"model"`
	Positive    PreNode `json:"positive"`
	Negative    PreNode `json:"negative"`
	LatentImage PreNode `json:"latent_image"`
}

func (i KSampler) Build() Node {
	return Node{
		ClassType: "KSampler",
		Inputs:    i,
	}
}

type CLIPTextEncode struct {
	Text string  `json:"text"`
	CLIP PreNode `json:"clip"`
}

func (i CLIPTextEncode) Build() Node {
	return Node{
		Inputs:    i,
		ClassType: "CLIPTextEncode",
	}
}

type PreviewImage struct {
	Images PreNode `json:"images"`
}

func (i PreviewImage) Build() Node {
	return Node{
		ClassType: "PreviewImage",
		Inputs:    i,
	}
}

type CLIPSetLastLayer struct {
	StopAtCLIPLayer int     `json:"stop_at_clip_layer"`
	CLIP            PreNode `json:"clip"`
}

func (i CLIPSetLastLayer) Build() Node {
	return Node{
		ClassType: "CLIPSetLastLayer",
		Inputs:    i,
	}
}

type ImageScale struct {
	UpscaleMethod string  `json:"upscale_method"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	Crop          string  `json:"crop"`
	Image         PreNode `json:"image"`
}

func (i ImageScale) Build() Node {
	return Node{
		ClassType: "ImageScale",
		Inputs:    i,
	}
}

type LatentUpscale struct {
	UpscaleMethod string  `json:"upscale_method"`
	Width         int     `json:"width"`
	Height        int     `json:"height"`
	Crop          string  `json:"crop"`
	Samples       PreNode `json:"samples"`
}

func (i LatentUpscale) Build() Node {
	return Node{
		ClassType: "LatentUpscale",
		Inputs:    i,
	}
}

type ImageScaleBy struct {
	UpscaleMethod string  `json:"upscale_method"`
	ScaleBy       float64 `json:"scale_by"`
	Image         PreNode `json:"image"`
}

func (i ImageScaleBy) Build() Node {
	return Node{
		ClassType: "ImageScaleBy",
		Inputs:    i,
	}
}

type VAEEncode struct {
	Pixels PreNode `json:"pixels"`
	VAE    PreNode `json:"vae"`
}

func (i VAEEncode) Build() Node {
	return Node{
		ClassType: "VAEEncode",
		Inputs:    i,
	}
}

type ImageUpscaleWithModel struct {
	UpscaleModel PreNode `json:"upscale_model"`
	Image        PreNode `json:"image"`
}

func (i ImageUpscaleWithModel) Build() Node {
	return Node{
		ClassType: "ImageUpscaleWithModel",
		Inputs:    i,
	}
}

type UpscaleModelLoader struct {
	ModelName string `json:"model_name"`
}

func (i UpscaleModelLoader) Build() Node {
	return Node{
		ClassType: "UpscaleModelLoader",
		Inputs:    i,
	}
}

type CheckpointLoaderSimple struct {
	Checkpoint string `json:"ckpt_name"`
}

func (i CheckpointLoaderSimple) Build() Node {
	return Node{
		ClassType: "CheckpointLoaderSimple",
		Inputs:    i,
	}
}

type EmptyLatentImage struct {
	Width     int `json:"width"`
	Height    int `json:"height"`
	BatchSize int `json:"batch_size"`
}

func (i EmptyLatentImage) Build() Node {
	return Node{
		ClassType: "EmptyLatentImage",
		Inputs:    i,
	}
}

type VAELoader struct {
	VAEName string `json:"vae_name"`
}

func (i VAELoader) Build() Node {
	return Node{
		ClassType: "VAELoader",
		Inputs:    i,
	}
}

type VAEDecode struct {
	Samples PreNode `json:"samples"`
	VAE     PreNode `json:"vae"`
}

func (i VAEDecode) Build() Node {
	return Node{
		ClassType: "VAEDecode",
		Inputs:    i,
	}
}
