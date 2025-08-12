package error

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComfyUIError(t *testing.T) {
	msg := json.RawMessage(`{"test":"1"}`)
	coreError := ComfyUIError{Message: msg, IsOOM: true}

	wrappedError1 := fmt.Errorf("wrapped 1: %w", coreError)
	wrappedError2 := fmt.Errorf("wrapped 2: %w", wrappedError1)

	var cErr ComfyUIError
	if errors.As(wrappedError2, &cErr) {
		assert.Equal(t, msg, cErr.Message)
	} else {
		t.Errorf("expect ComfyUIError")
	}

	wrappedError3 := errors.Join(errors.New("others"), wrappedError2)
	var cErr2 ComfyUIError
	if errors.As(wrappedError3, &cErr2) {
		assert.Equal(t, msg, cErr2.Message)
		assert.Equal(t, true, cErr2.IsOOM)
	} else {
		t.Errorf("expect ComfyUIError")
	}
}

func TestMarshalJSON(t *testing.T) {
	marshalErr := func(err ComfyUIError) ([]byte, error) {
		return json.Marshal(map[string]any{
			"err": err.Message,
		})
	}
	tests := []struct {
		err     ComfyUIError
		expect  string
		wantErr bool
	}{
		{
			err:    ComfyUIError{},
			expect: `{"err":null}`,
		},
		{
			// empty message, but not nil, should be error
			err:     ComfyUIError{Message: json.RawMessage{}},
			wantErr: true,
		},
		{
			err:    ComfyUIError{Message: json.RawMessage(`{"test":"1"}`)},
			expect: `{"err":{"test":"1"}}`,
		},
	}

	for _, test := range tests {
		json, err := marshalErr(test.err)
		if test.wantErr {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expect, string(json))
		}
	}
}
