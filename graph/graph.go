/*
graph/graph.go

convert from graph workflow to prompt

*/
package graph

import (
	"encoding/json"
	"fmt"
)

// Graph format structures
type GraphData struct {
	LastNodeID int    `json:"last_node_id"`
	LastLinkID int    `json:"last_link_id"`
	Nodes      []Node `json:"nodes"`
	Links      []Link `json:"links"`
}

type Node struct {
	ID            string   `json:"id"`
	Type          string   `json:"type"`
	Inputs        []Input  `json:"inputs"`
	Outputs       []Output `json:"outputs"`
	Title         *string  `json:"title"`
	WidgetsValues any      `json:"widgets_values"`
}

func (n *Node) UnmarshalJSON(p []byte) error {
	type Alias0 Node
	type Alias struct {
		Alias0
		ID int `json:"id"`
	}
	var alias Alias
	if err := json.Unmarshal(p, &alias); err != nil {
		return err
	}
	*n = Node(alias.Alias0)
	n.ID = fmt.Sprintf("%d", alias.ID)
	return nil
}

type Input struct {
	Name   string         `json:"name"`
	Type   string         `json:"type"`
	Link   string         `json:"link"`
	Widget map[string]any `json:"widget"` // means this input has a widget
	Label  string         `json:"label"`
}

func (i *Input) UnmarshalJSON(p []byte) error {
	type Alias0 Input
	type Alias struct {
		Alias0
		Link *int `json:"link"`
	}
	var alias Alias
	if err := json.Unmarshal(p, &alias); err != nil {
		return err
	}
	*i = Input(alias.Alias0)
	if alias.Link != nil {
		i.Link = fmt.Sprintf("%d", *alias.Link)
	}
	return nil
}

type Output struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Links     []int  `json:"links"`
	SlotIndex int    `json:"slot_index"`
	Label     string `json:"label"`
}

// 6 个元素的数组，分别表示：ID, FromNode, FromOutput, ToNode, ToInput, Type
type Link struct {
	ID         string
	FromNode   string
	FromOutput int
	ToNode     string
	ToOutput   int
	Type       string
}

func (l *Link) UnmarshalJSON(p []byte) error {
	var link []any
	if err := json.Unmarshal(p, &link); err != nil {
		return err
	}
	if len(link) < 6 {
		return fmt.Errorf("link must have at least 6 elements")
	}
	l.ID = fmt.Sprintf("%v", link[0])
	l.FromNode = fmt.Sprintf("%v", link[1])
	l.FromOutput = int(link[2].(float64))
	l.ToNode = fmt.Sprintf("%v", link[3])
	l.ToOutput = int(link[4].(float64))
	l.Type = fmt.Sprintf("%v", link[5])
	return nil
}

// API format structures
type APIPrompt map[string]NodeData

type NodeData struct {
	Inputs    map[string]interface{} `json:"inputs"`
	ClassType string                 `json:"class_type"`
	Meta      NodeMeta               `json:"_meta"`
}

type NodeMeta struct {
	Title string `json:"title"`
}
