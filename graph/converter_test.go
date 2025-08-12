package graph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGraphConverter_Convert(t *testing.T) {
	// Setup test data
	testDataDir := "test"
	objectInfoDir := filepath.Join(testDataDir, "object_info")

	// Create file-based object info fetcher
	fetcher, err := NewFileObjectInfoFetcher(objectInfoDir)
	assert.NoError(t, err)

	// Create graph converter
	converter := NewGraphConverter(fetcher)

	tests := []string{
		"primitive",
		"txt2img",
		"workflow_reroute",
	}
	for _, test := range tests {
		t.Run(test, func(t *testing.T) {
			graphFilename := test + ".json"
			apiFilename := test + "_api.json"

			// Read test input graph
			graphData, err := os.ReadFile(filepath.Join(testDataDir, graphFilename))
			assert.NoError(t, err)

			// Read expected API output
			expectedAPI, err := os.ReadFile(filepath.Join(testDataDir, apiFilename))
			assert.NoError(t, err)

			// Convert graph data to raw message
			var rawMessage json.RawMessage = graphData

			// Convert graph
			result, err := converter.Convert(rawMessage)
			assert.NoError(t, err)

			// Compare results
			var expected, actual map[string]interface{}

			err = json.Unmarshal(expectedAPI, &expected)
			assert.NoError(t, err)

			err = json.Unmarshal(result, &actual)
			assert.NoError(t, err)

			// Deep compare the results
			assert.Equal(t, expected, actual, "Conversion result should match expected output")
		})
	}
}
