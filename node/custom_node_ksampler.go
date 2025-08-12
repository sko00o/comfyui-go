package node

// Extension: ComfyUI-Inspire-Pack (https://github.com/ltdrdata/ComfyUI-Inspire-Pack)
// Node: KSampler //Inspire

var _ Builder = (*KSamplerA1111)(nil)
var _ Builder = (*KSamplerInspire)(nil)

type KSamplerA1111 KSampler

type KSamplerInspire struct {
	KSamplerA1111
	NoiseMode         string `json:"noise_mode"`
	BatchSeedMode     string `json:"batch_seed_mode"`
	VariationSeed     int    `json:"variation_seed"`
	VariationStrength int    `json:"variation_strength"`
	VariationMethod   string `json:"variation_method"`
}

func (i KSamplerA1111) Build() Node {
	newInputs := KSamplerInspire{
		KSamplerA1111:     i,
		NoiseMode:         "GPU(=A1111)",
		BatchSeedMode:     "incremental",
		VariationSeed:     0,
		VariationStrength: 0,
		VariationMethod:   "linear",
	}
	return Node{
		ClassType: "KSampler //Inspire",
		Inputs:    newInputs,
	}
}
