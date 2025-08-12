package node

// Extension: smZNodes (https://github.com/shiimizu/ComfyUI_smZNodes)
// Node: smZ CLIPTextEncode

var _ Builder = (*CLIPTextEncodeA1111)(nil)

type CLIPTextEncodeA1111 CLIPTextEncode

func (i CLIPTextEncodeA1111) Build() Node {
	newInputs := map[string]any{
		"text": i.Text,
		"clip": i.CLIP,
		// for A1111 reproduce
		"parser":                          "A1111",
		"mean_normalization":              true,
		"multi_conditioning":              false,
		"use_old_emphasis_implementation": false,
		"with_SDXL":                       false,
		"ascore":                          6,
		"width":                           1024,
		"height":                          1024,
		"crop_w":                          0,
		"crop_h":                          0,
		"target_width":                    1024,
		"target_height":                   1024,
		"text_g":                          "",
		"text_l":                          "",
		"smZ_steps":                       1,
	}
	return Node{
		ClassType: "smZ CLIPTextEncode",
		Inputs:    newInputs,
	}
}
