package prompt

const (
	IDLoraLoader = "55"

	IDKSampler               = "101"
	IDCLIPTextEncodePositive = "102"
	IDCLIPTextEncodeNegative = "103"
	IDSaveImage              = "104"
	IDVAEDecode              = "105"
	IDVAELoader              = "106"
	IDCheckpointLoader       = "107"
	IDCLIPSetLastLayer       = "108"
	IDEmptyLatentImage       = "109"

	// Upscale
	IDUpscaleModelLoader    = "211"
	IDImageUpscaleWithModel = "212"
	IDVAEEncode             = "213"
	IDImageScale            = "214"
	IDKSamplerUpscale       = "215"
	IDVAEDecodeUpscale      = "216"
	IDLatentScale           = "217"

	// SD3.5
	IDTripleCLIPLoader              = "18"
	IDModelSamplingSD3              = "19"
	IDConditioningZeroOut           = "20"
	IDConditioningSetTimestepRange1 = "21"
	IDConditioningSetTimestepRange2 = "22"
	IDConditioningCombine           = "23"
	IDEmptySD3LatentImage           = "24"

	// Flux
	FluxIDRandomNoise          = "60"
	FluxIDPerpNegGuider        = "61"
	FluxIDKSamplerSelect       = "62"
	FluxIDBasicScheduler       = "63"
	FluxIDModelSamplingFlux    = "64"
	FluxIDBasicGuider          = "65"
	FluxIDUNETLoader           = "66"
	FluxIDFluxPositiveGuidance = "67"
	FluxIDFluxNegativeGuidance = "68"
	FluxIDEmptyCLIPTextEncode  = "69"
	FluxIDDualCLIPLoader       = "70"
)
