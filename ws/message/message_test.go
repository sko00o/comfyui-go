package message

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestFileInfo_JSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		want        FileInfo
		modify      func(*FileInfo)
		afterModify string
		wantErr     bool
	}{
		{
			name:  "basic fields",
			input: `{"filename":"test.png","subfolder":"outputs","type":"image"}`,
			want: FileInfo{
				Filename:  "test.png",
				Subfolder: "outputs",
				Type:      "image",
			},
		},
		{
			name:  "with extra fields",
			input: `{"filename":"test.png","subfolder":"outputs","type":"image","extra":"value","num":123}`,
			want: FileInfo{
				Filename:  "test.png",
				Subfolder: "outputs",
				Type:      "image",
			},
		},
		{
			name:  "modify fields",
			input: `{"filename":"old.png","subfolder":"old","type":"image","extra":"value"}`,
			want: FileInfo{
				Filename:  "old.png",
				Subfolder: "old",
				Type:      "image",
			},
			modify: func(f *FileInfo) {
				f.Filename = "new.png"
				f.Subfolder = "new"
				f.Type = "mask"
			},
			afterModify: `{"filename":"new.png","subfolder":"new","type":"mask","extra":"value"}`,
		},
		{
			name:    "invalid json",
			input:   `{"filename":123}`, // filename should be string
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got FileInfo
			err := json.Unmarshal([]byte(tt.input), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileInfo.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}

			// Check basic fields
			if got.Filename != tt.want.Filename {
				t.Errorf("Filename = %v, want %v", got.Filename, tt.want.Filename)
			}
			if got.Subfolder != tt.want.Subfolder {
				t.Errorf("Subfolder = %v, want %v", got.Subfolder, tt.want.Subfolder)
			}
			if got.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", got.Type, tt.want.Type)
			}

			// Apply modifications if any
			if tt.modify != nil {
				tt.modify(&got)
			}

			// Test marshaling
			data, err := json.Marshal(got)
			if err != nil {
				t.Errorf("FileInfo.MarshalJSON() error = %v", err)
				return
			}

			// Unmarshal both to map to compare
			var gotMap, wantMap map[string]interface{}
			if err := json.Unmarshal(data, &gotMap); err != nil {
				t.Errorf("Failed to unmarshal result: %v", err)
				return
			}

			wantOutput := tt.input
			if tt.afterModify != "" {
				wantOutput = tt.afterModify
			}
			if err := json.Unmarshal([]byte(wantOutput), &wantMap); err != nil {
				t.Errorf("Failed to unmarshal input: %v", err)
				return
			}

			if !reflect.DeepEqual(gotMap, wantMap) {
				t.Errorf("FileInfo JSON roundtrip = %v, want %v", gotMap, wantMap)
			}
		})
	}
}

func TestDataProgress_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    DataProgress
		wantErr bool
	}{
		{
			name:  "normal int values",
			input: `{"prompt_id":"test123","node":"test_node","value":50,"max":100}`,
			want: DataProgress{
				DataExecuting: DataExecuting{
					ExecutionBase: ExecutionBase{
						PromptID: "test123",
					},
					Node: stringPtr("test_node"),
				},
				Value: 50,
				Max:   100,
			},
		},
		{
			name:  "float64 values",
			input: `{"prompt_id":"test123","value":50.0,"max":100.0}`,
			want: DataProgress{
				DataExecuting: DataExecuting{
					ExecutionBase: ExecutionBase{
						PromptID: "test123",
					},
				},
				Value: 50,
				Max:   100,
			},
		},
		{
			name:  "mixed float and int",
			input: `{"prompt_id":"test123","value":25.5,"max":200}`,
			want: DataProgress{
				DataExecuting: DataExecuting{
					ExecutionBase: ExecutionBase{
						PromptID: "test123",
					},
				},
				Value: 25,
				Max:   200,
			},
		},
		{
			name:  "int64 values",
			input: `{"prompt_id":"test123","value":75,"max":150}`,
			want: DataProgress{
				DataExecuting: DataExecuting{
					ExecutionBase: ExecutionBase{
						PromptID: "test123",
					},
				},
				Value: 75,
				Max:   150,
			},
		},
		{
			name:    "invalid value type",
			input:   `{"prompt_id":"test123","value":"not_a_number","max":100}`,
			wantErr: true,
		},
		{
			name:    "invalid max type",
			input:   `{"prompt_id":"test123","value":50,"max":"not_a_number"}`,
			wantErr: true,
		},
		{
			name:    "missing value field",
			input:   `{"prompt_id":"test123","max":100}`,
			wantErr: true,
		},
		{
			name:    "missing max field",
			input:   `{"prompt_id":"test123","value":50}`,
			wantErr: true,
		},
		{
			name:  "zero values",
			input: `{"prompt_id":"test123","value":0,"max":0}`,
			want: DataProgress{
				DataExecuting: DataExecuting{
					ExecutionBase: ExecutionBase{
						PromptID: "test123",
					},
				},
				Value: 0,
				Max:   0,
			},
		},
		{
			name:  "large numbers",
			input: `{"prompt_id":"test123","value":999999,"max":999999}`,
			want: DataProgress{
				DataExecuting: DataExecuting{
					ExecutionBase: ExecutionBase{
						PromptID: "test123",
					},
				},
				Value: 999999,
				Max:   999999,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got DataProgress
			err := json.Unmarshal([]byte(tt.input), &got)

			if (err != nil) != tt.wantErr {
				t.Errorf("DataProgress.UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			// Check fields
			if got.PromptID != tt.want.PromptID {
				t.Errorf("PromptID = %v, want %v", got.PromptID, tt.want.PromptID)
			}
			if got.Value != tt.want.Value {
				t.Errorf("Value = %v, want %v", got.Value, tt.want.Value)
			}
			if got.Max != tt.want.Max {
				t.Errorf("Max = %v, want %v", got.Max, tt.want.Max)
			}
			if !reflect.DeepEqual(got.Node, tt.want.Node) {
				t.Errorf("Node = %v, want %v", got.Node, tt.want.Node)
			}
			if !reflect.DeepEqual(got.DisplayNode, tt.want.DisplayNode) {
				t.Errorf("DisplayNode = %v, want %v", got.DisplayNode, tt.want.DisplayNode)
			}
		})
	}
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
