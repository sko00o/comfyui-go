package node

// Extension: LoRA Tag Loader for ComfyUI (https://github.com/badjeff/comfyui_lora_tag_loader)
// Node: LoraTagLoader

type LoraTagLoader struct {
	Text  string  `json:"text"`
	Model PreNode `json:"model"`
	CLIP  PreNode `json:"clip"`
}

func (i LoraTagLoader) Build() Node {
	return Node{
		ClassType: "LoraTagLoader",
		Inputs:    i,
	}
}
