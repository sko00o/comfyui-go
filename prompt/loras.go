package prompt

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/sko00o/comfyui-go/node"
)

type Loras []Lora

func (l Loras) String() string {
	loras := make([]string, 0, len(l))
	for _, i := range l {
		loras = append(loras, i.String())
	}
	return strings.Join(loras, ",")
}

func (l Loras) Build(baseID string, model, clip node.PreNode) (wf Prompt, linkModel node.PreNode, linkCLIP node.PreNode) {
	wf = make(Prompt)
	if len(l) == 0 {
		return wf, model, clip
	}
	currID := baseID
	linkModel = node.PreNode{ID: baseID, Argc: 0}
	linkCLIP = node.PreNode{ID: baseID, Argc: 1}
	for i := len(l) - 1; i >= 0; i-- {
		lora := l[i]
		nextID := fmt.Sprintf("%s:%d", baseID, i)
		currModel := node.PreNode{ID: nextID, Argc: 0}
		currCLIP := node.PreNode{ID: nextID, Argc: 1}
		if i == 0 {
			currModel = model
			currCLIP = clip
		}
		wf[currID] = node.GeneralNode{
			ClassType: "LoraLoader",
			Inputs: map[string]any{
				"lora_name":      lora.LoraName,
				"strength_model": lora.StrengthModel,
				"strength_clip":  lora.StrengthCLIP,
				"model":          currModel,
				"clip":           currCLIP,
			},
		}
		currID = nextID
	}
	return wf, linkModel, linkCLIP
}

type Lora struct {
	LoraName      string  `json:"lora_name"`
	StrengthModel float64 `json:"strength_model"`
	StrengthCLIP  float64 `json:"strength_clip"`
}

func (i Lora) String() string {
	loraText := "<lora:" + i.LoraName
	if i.StrengthModel > 0 {
		loraText += ":" + strconv.FormatFloat(i.StrengthModel, 'f', -1, 64)
	} else {
		loraText += ":1"
	}
	if i.StrengthCLIP > 0 {
		loraText += ":" + strconv.FormatFloat(i.StrengthCLIP, 'f', -1, 64)
	} else {
		loraText += ":1"
	}
	loraText += ">"
	return loraText
}
