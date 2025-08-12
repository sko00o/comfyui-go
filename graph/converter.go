package graph

import (
	"encoding/json"
	"fmt"
	"reflect"

	nd "github.com/sko00o/comfyui-go/node"
)

// GraphConverter 用于转换图形的结构体
type GraphConverter struct {
	fetcher ObjectInfoFetcher
}

// ObjectInfoFetcher 定义获取 object_info 的接口
type ObjectInfoFetcher interface {
	FetchNodeInfo(nodeType string) (*NodeInfo, error)
}

func NewGraphConverter(fetcher ObjectInfoFetcher) *GraphConverter {
	return &GraphConverter{
		fetcher: fetcher,
	}
}

// Convert 转换图形数据
func (c *GraphConverter) Convert(graph json.RawMessage) ([]byte, error) {
	var graphData GraphData
	if err := json.Unmarshal(graph, &graphData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal graph data: %v", err)
	}

	nodeMap := make(map[string]Node)
	for _, node := range graphData.Nodes {
		nodeMap[node.ID] = node
	}

	apiPrompt := make(APIPrompt)

	// 创建节点映射表，用于追踪所有需要重定向的节点（包括 Reroute 和 GetNode）
	nodeRedirectMap := make(map[string]*nd.PreNode)

	// 第一遍遍历：收集所有 Reroute 节点的连接信息
	variableMap := make(map[string]*nd.PreNode)
	primitiveMap := make(map[string]any)
	linkMap := make(map[string]Link)
	for _, link := range graphData.Links {
		linkMap[link.ID] = link

		if node, exists := nodeMap[link.ToNode]; exists {
			switch node.Type {
			case "Reroute", "Reroute (rgthree)":
				nodeRedirectMap[link.ToNode] = &nd.PreNode{
					ID:   link.FromNode,
					Argc: link.FromOutput,
				}
			case "SetNode":
				nodeRedirectMap[link.ToNode] = &nd.PreNode{
					ID:   link.FromNode,
					Argc: link.FromOutput,
				}
				if node.WidgetsValues != nil {
					if v, ok := node.WidgetsValues.([]interface{}); ok && len(v) > 0 {
						varName := v[0].(string)
						variableMap[varName] = &nd.PreNode{
							ID:   link.FromNode,
							Argc: link.FromOutput,
						}
					}
				}
			}
		}

		if node, exists := nodeMap[link.FromNode]; exists {
			switch node.Type {
			case "PrimitiveNode":
				primitiveMap[node.ID] = node.WidgetsValues
			}
		}
	}

	// 收集所有 GetNode 的映射关系
	for _, node := range graphData.Nodes {
		if node.Type == "GetNode" {
			if node.WidgetsValues != nil {
				if v, ok := node.WidgetsValues.([]interface{}); ok && len(v) > 0 {
					varName := v[0].(string)
					if source, exists := variableMap[varName]; exists {
						nodeRedirectMap[node.ID] = source
					}
				}
			}
		}
	}

	// 第四遍遍历：处理所有节点
	for _, node := range graphData.Nodes {
		if node.Type == "Note" || node.Type == "Note Plus (mtb)" || node.Type == "Note _O" ||
			node.Type == "Reroute" || node.Type == "Reroute (rgthree)" ||
			node.Type == "SetNode" ||
			node.Type == "GetNode" ||
			node.Type == "PrimitiveNode" {
			continue
		}

		// 获取节点信息
		nodeInfo, err := c.fetcher.FetchNodeInfo(node.Type)
		if err != nil {
			return nil, fmt.Errorf("failed to get node info for %s: %v", node.Type, err)
		}

		// 创建输入映射
		inputs := make(map[string]interface{})

		// 1. 先处理有 link 的输入
		linkedInputs := make(map[string]bool)
		hasWidget := make(map[string]bool)
		for _, input := range node.Inputs {
			if link, ok := linkMap[input.Link]; ok {
				fromNode, fromOutput := link.FromNode, link.FromOutput
				last := findOriginalSource(nodeRedirectMap, fromNode)
				if last != nil {
					fromNode, fromOutput = last.ID, last.Argc
				}
				if primitive, exists := primitiveMap[fromNode]; exists {
					if v, ok := primitive.([]interface{}); ok && len(v) > 0 {
						inputs[input.Name] = v[0]
					}
				} else {
					inputs[input.Name] = []any{
						fromNode,
						fromOutput,
					}
				}
				linkedInputs[input.Name] = true
			}
			if input.Widget != nil {
				hasWidget[input.Name] = true
			}
		}

		// 2. 处理 widgets values 和默认值
		if nodeInfo.Input.Optional != nil {
			// 先设置所有可选参数的默认值
			for paramName, paramDef := range nodeInfo.Input.Optional {
				if len(paramDef) > 1 {
					if v, ok := paramDef[1].(map[string]interface{}); ok {
						if defaultValue, exists := v["default"]; exists {
							inputs[paramName] = defaultValue
						}
					}
				}
			}
		}

		if node.WidgetsValues != nil {
			var currentIndex int

			// 检查 WidgetsValues 的类型
			switch widgetsValue := node.WidgetsValues.(type) {
			case []interface{}: // 处理数组类型
				currentIndex = 0
				widgetsValueSlice := reflect.ValueOf(widgetsValue)

				// 按顺序处理所有输入参数
				for _, paramName := range nodeInfo.InputOrder.Required {
					if linkedInputs[paramName] {
						if hasWidget[paramName] {
							if paramName == "seed" || paramName == "noise_seed" {
								currentIndex += 2
							} else {
								currentIndex += 1
							}
						}
						continue
					}

					if currentIndex >= widgetsValueSlice.Len() {
						break
					}

					paramDef, exists := nodeInfo.Input.Required[paramName]
					if !exists {
						continue
					}

					currentIndex = processParamSlice(inputs, paramName, paramDef, widgetsValueSlice, currentIndex)
				}

				// 处理可选参数
				for _, paramName := range nodeInfo.InputOrder.Optional {
					if linkedInputs[paramName] {
						if hasWidget[paramName] {
							currentIndex += 1
						}
						continue
					}

					if currentIndex >= widgetsValueSlice.Len() {
						break
					}

					paramDef, exists := nodeInfo.Input.Optional[paramName]
					if !exists {
						continue
					}

					currentIndex = processParamSlice(inputs, paramName, paramDef, widgetsValueSlice, currentIndex)
				}

			case map[string]interface{}: // 处理 map 类型
				// special case in VHS_VideoCombine
				for _, s := range []string{
					"pix_fmt",
					"crf",
					"save_metadata",
				} {
					if value, exists := widgetsValue[s]; exists {
						inputs[s] = value
					}
				}

				// 直接处理所有必需参数
				for paramName, paramDef := range nodeInfo.Input.Required {
					if linkedInputs[paramName] {
						continue
					}
					if value, exists := widgetsValue[paramName]; exists {
						processParamMap(inputs, paramName, paramDef, value)
					}
				}

				// 处理所有可选参数
				for paramName, paramDef := range nodeInfo.Input.Optional {
					if linkedInputs[paramName] {
						continue
					}
					if value, exists := widgetsValue[paramName]; exists {
						processParamMap(inputs, paramName, paramDef, value)
					}
				}
			}
		}

		title := nodeInfo.DisplayName
		if node.Title != nil {
			title = *node.Title
		}

		nodeData := NodeData{
			Inputs:    inputs,
			ClassType: node.Type,
			Meta: NodeMeta{
				Title: title,
			},
		}

		apiPrompt[node.ID] = nodeData
	}

	// 转换为 JSON 字符串
	result, err := json.MarshalIndent(apiPrompt, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal API prompt: %v", err)
	}

	return result, nil
}

// 添加新的辅助函数，用于追踪 Reroute 链的源头
func findOriginalSource(nodeRedirectMap map[string]*nd.PreNode, nodeID string) *nd.PreNode {
	last, exists := nodeRedirectMap[nodeID]
	if !exists {
		return nil
	}

	for {
		last1 := findOriginalSource(nodeRedirectMap, last.ID)
		if last1 == nil {
			break
		}
		if last1.ID != last.ID {
			nodeRedirectMap[nodeID] = last1
			last = last1
		}
	}
	return last
}

// InputDef 定义输入参数的结构: [type, {options}]
type InputDef []interface{}

// NodeInfo 存储节点的输入参数定义
type NodeInfo struct {
	Input struct {
		Required map[string]InputDef `json:"required"`
		Optional map[string]InputDef `json:"optional"`
	} `json:"input"`
	InputOrder struct {
		Required []string `json:"required"`
		Optional []string `json:"optional"`
	} `json:"input_order"`
	DisplayName string `json:"display_name"`
}

func processParamSlice(inputs map[string]interface{}, paramName string, paramDef InputDef, widgetsValue reflect.Value, index int) int {
	widget := widgetsValue.Index(index).Interface()
	if widget == nil {
		return index + 1
	}

	// param option has "image_upload": true
	if len(paramDef) > 1 {
		// convert to map[string]interface{}
		paramOption := paramDef[1].(map[string]interface{})
		if paramOption["image_upload"] == true {
			// add upload parameter
			inputs["upload"] = paramName
		}
	}

	// 如果是数组类型（比如选项列表），直接使用 widget 值
	if reflect.TypeOf(paramDef[0]).Kind() == reflect.Slice {
		inputs[paramName] = widget
		return index + 1
	}

	paramType, ok := paramDef[0].(string)
	if !ok {
		return index + 1
	}

	// 处理基本类型
	switch paramType {
	case "INT":
		if v, ok := widget.(float64); ok {
			inputs[paramName] = int(v)
		}
		if paramName == "seed" || paramName == "noise_seed" {
			// seed 参数特殊处理：跳过 control_after_generate 选项的 widget 值
			return index + 2
		}
	case "FLOAT":
		if v, ok := widget.(float64); ok {
			inputs[paramName] = v
		}
	case "STRING":
		if v, ok := widget.(string); ok {
			inputs[paramName] = v
		}
	case "BOOLEAN":
		if v, ok := widget.(bool); ok {
			inputs[paramName] = v
		} else {
			// 可能存在错误类型，直接填
			inputs[paramName] = widget
		}
	default:
		// 非基础类型都不跳过
		return index
	}

	return index + 1
}

// 新增处理 map 类型的函数
func processParamMap(inputs map[string]interface{}, paramName string, paramDef InputDef, value interface{}) {
	if value == nil {
		return
	}

	// param option has "image_upload": true
	if len(paramDef) > 1 {
		paramOption := paramDef[1].(map[string]interface{})
		if paramOption["image_upload"] == true {
			inputs["upload"] = paramName
		}
	}

	// 如果是数组类型（比如选项列表），直接使用值
	if reflect.TypeOf(paramDef[0]).Kind() == reflect.Slice {
		inputs[paramName] = value
		return
	}

	paramType, ok := paramDef[0].(string)
	if !ok {
		return
	}

	// 处理基本类型
	switch paramType {
	case "INT":
		switch v := value.(type) {
		case float64:
			inputs[paramName] = int(v)
		case int:
			inputs[paramName] = v
		}
	case "FLOAT":
		if v, ok := value.(float64); ok {
			inputs[paramName] = v
		}
	case "STRING":
		if v, ok := value.(string); ok {
			inputs[paramName] = v
		}
	case "BOOLEAN":
		if v, ok := value.(bool); ok {
			inputs[paramName] = v
		}
	}
}
