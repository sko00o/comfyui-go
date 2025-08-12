package node

// Extensions: websocket_image_save.py
// Node: SaveImageWebsocket
//
// The output are binary images on the websocket with an 8 byte header indicating
// the type of binary message (first 4 bytes) and the image format (next 4 bytes).

var _ Builder = (*SaveImageWebsocket)(nil)

type SaveImageWebsocket struct {
	Images PreNode `json:"images"`

	EnableMetadata bool `json:"enable_metadata,omitempty"`
}

func (i SaveImageWebsocket) Build() Node {
	classType := "SaveImageWebsocket"
	if i.EnableMetadata {
		classType = "SaveImageWithPromptsWebsocket"
	}
	return Node{
		ClassType: classType,
		Inputs:    i,
	}
}
