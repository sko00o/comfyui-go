package message

import (
	"encoding"
	"encoding/binary"
	"fmt"
)

type EventType uint32

const (
	PreviewImage EventType = 1 + iota
	UnencodedPreviewImage
)

type ImageType uint32

const (
	JPEG ImageType = iota + 1
	PNG
)

func (t ImageType) Ext() string {
	switch t {
	case JPEG:
		return ".jpeg"
	case PNG:
		return ".png"
	default:
		return ""
	}
}

func (t ImageType) ContentType() string {
	switch t {
	case JPEG:
		return "image/jpeg"
	case PNG:
		return "image/png"
	default:
		return ""
	}
}

type BinaryMessage struct {
	Type EventType
	Data encoding.BinaryUnmarshaler
}

func (m *BinaryMessage) UnmarshalBinary(message []byte) error {
	if len(message) < 4 {
		return fmt.Errorf("length too short")
	}

	m.Type = EventType(binary.BigEndian.Uint32(message[:4]))
	switch m.Type {
	case PreviewImage:
		m.Data = &DataImage{}
	default:
		return fmt.Errorf("unknown type %v", m.Type)
	}

	buffer := message[4:]
	return m.Data.UnmarshalBinary(buffer)
}

type DataImage struct {
	Type ImageType
	Blob []byte
}

func (m *DataImage) UnmarshalBinary(buffer []byte) error {
	if len(buffer) < 4 {
		return fmt.Errorf("image data length too short")
	}

	m.Type = ImageType(binary.BigEndian.Uint32(buffer[:4]))
	switch m.Type {
	case JPEG, PNG:
	default:
		return fmt.Errorf("unknow image type %v", m.Type)
	}

	m.Blob = buffer[4:]
	return nil
}
