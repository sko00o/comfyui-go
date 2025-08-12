package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sko00o/comfyui-go"
	"github.com/sko00o/comfyui-go/prompt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWorkflowIntegration tests workflow integration functionality
func TestWorkflowIntegration(t *testing.T) {
	// Skip workflow tests - requires GPU for inference
	t.Skip("Skipping workflow integration test - requires GPU for inference tasks")

	// Start ComfyUI container
	ctx := context.Background()
	container, err := startComfyUIContainer(ctx)
	require.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	// Get container port
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "8188")
	require.NoError(t, err)

	// Wait for service to start
	time.Sleep(10 * time.Second)

	// Create ComfyUI client
	client, err := comfyui.New(comfyui.Config{
		Endpoint: fmt.Sprintf("http://%s:%s", host, port.Port()),
		Timeout:  30 * time.Second,
	})
	require.NoError(t, err)

	// Test different types of workflows
	t.Run("TestA1111Workflow", func(t *testing.T) {
		testA1111Workflow(t, client)
	})

	t.Run("TestFluxWorkflow", func(t *testing.T) {
		testFluxWorkflow(t, client)
	})

	t.Run("TestSD3Workflow", func(t *testing.T) {
		testSD3Workflow(t, client)
	})
}

// testA1111Workflow tests A1111 workflow
func testA1111Workflow(t *testing.T, client *comfyui.Client) {
	// Create A1111 workflow
	wf := &prompt.A1111Base{
		Base: prompt.Base{
			Checkpoint:     "v1-5-pruned.ckpt",
			PositivePrompt: "a beautiful landscape",
			NegativePrompt: "blurry, low quality",
			ImageWidth:     512,
			ImageHeight:    512,
			BatchSize:      1,
			Seed:           42,
			Steps:          20,
			CFG:            7.0,
			SamplerName:    "euler",
			Scheduler:      "normal",
			Type:           "a1111",
		},
		CLIPSkip: 1,
	}

	// Get workflow data
	workflowData := wf.Build()
	require.NotNil(t, workflowData)

	// Convert to map[string]any
	workflowMap := make(map[string]any)
	for k, v := range workflowData {
		workflowMap[k] = v
	}

	// Submit workflow to ComfyUI
	resp, err := client.Prompt(workflowMap)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.PromptID)
	assert.GreaterOrEqual(t, resp.Number, 0)

	t.Logf("Submitted A1111 workflow with prompt ID: %s", resp.PromptID)
}

// testFluxWorkflow tests Flux workflow
func testFluxWorkflow(t *testing.T, client *comfyui.Client) {
	// Create Flux workflow
	wf := &prompt.FluxBase{
		Base: prompt.Base{
			Checkpoint:     "flux.safetensors",
			PositivePrompt: "a beautiful landscape",
			NegativePrompt: "blurry, low quality",
			ImageWidth:     512,
			ImageHeight:    512,
			BatchSize:      1,
			Seed:           42,
			Steps:          20,
			CFG:            7.0,
			SamplerName:    "euler",
			Scheduler:      "normal",
			Type:           "flux",
		},
		WeightDtype: "fp8_e5m2",
		VAEName:     "ae.safetensors",
	}

	// Get workflow data
	workflowData := wf.Build()
	require.NotNil(t, workflowData)

	// Convert to map[string]any
	workflowMap := make(map[string]any)
	for k, v := range workflowData {
		workflowMap[k] = v
	}

	// Submit workflow to ComfyUI
	resp, err := client.Prompt(workflowMap)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.PromptID)
	assert.GreaterOrEqual(t, resp.Number, 0)

	t.Logf("Submitted Flux workflow with prompt ID: %s", resp.PromptID)
}

// testSD3Workflow tests SD3 workflow
func testSD3Workflow(t *testing.T, client *comfyui.Client) {
	// Create SD3 workflow
	wf := &prompt.SD3Base{
		Base: prompt.Base{
			Checkpoint:     "sd3.safetensors",
			PositivePrompt: "a beautiful landscape",
			NegativePrompt: "blurry, low quality",
			ImageWidth:     512,
			ImageHeight:    512,
			BatchSize:      1,
			Seed:           42,
			Steps:          20,
			CFG:            7.0,
			SamplerName:    "euler",
			Scheduler:      "normal",
			Type:           "sd3",
		},
		VAEName: "",
	}

	// Get workflow data
	workflowData := wf.Build()
	require.NotNil(t, workflowData)

	// Convert to map[string]any
	workflowMap := make(map[string]any)
	for k, v := range workflowData {
		workflowMap[k] = v
	}

	// Submit workflow to ComfyUI
	resp, err := client.Prompt(workflowMap)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.PromptID)
	assert.GreaterOrEqual(t, resp.Number, 0)

	t.Logf("Submitted SD3 workflow with prompt ID: %s", resp.PromptID)
}

// TestWorkflowWithWebSocket tests workflow execution with WebSocket
func TestWorkflowWithWebSocket(t *testing.T) {
	// Skip workflow tests - requires GPU for inference
	t.Skip("Skipping workflow WebSocket test - requires GPU for inference tasks")

	// Start ComfyUI container
	ctx := context.Background()
	container, err := startComfyUIContainer(ctx)
	require.NoError(t, err)
	defer func() {
		if err := container.Terminate(ctx); err != nil {
			t.Logf("Failed to terminate container: %v", err)
		}
	}()

	// Get container port
	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "8188")
	require.NoError(t, err)

	// Wait for service to start
	time.Sleep(10 * time.Second)

	// Create ComfyUI client
	client, err := comfyui.New(comfyui.Config{
		Endpoint: fmt.Sprintf("http://%s:%s", host, port.Port()),
		Timeout:  30 * time.Second,
	})
	require.NoError(t, err)

	// Create WebSocket message handler
	handler := &workflowMessageHandler{
		executionStart: make(chan struct{}),
		executionEnd:   make(chan struct{}),
		progress:       make(chan int, 10),
	}

	// Start WebSocket processing
	wg, err := client.SimpleProcess("test-workflow-client", handler)
	require.NoError(t, err)
	defer wg.Wait()

	// Create simple workflow
	wf := &prompt.A1111Base{
		Base: prompt.Base{
			Checkpoint:     "v1-5-pruned.ckpt",
			PositivePrompt: "test image",
			NegativePrompt: "",
			ImageWidth:     256,
			ImageHeight:    256,
			BatchSize:      1,
			Seed:           42,
			Steps:          5, // Use fewer steps to speed up testing
			CFG:            7.0,
			SamplerName:    "euler",
			Scheduler:      "normal",
			Type:           "a1111",
		},
		CLIPSkip: 1,
	}

	// Get workflow data
	workflowData := wf.Build()
	require.NotNil(t, workflowData)

	// Convert to map[string]any
	workflowMap := make(map[string]any)
	for k, v := range workflowData {
		workflowMap[k] = v
	}

	// Submit workflow
	resp, err := client.Prompt(workflowMap)
	require.NoError(t, err)
	assert.NotEmpty(t, resp.PromptID)

	// Wait for execution to start
	select {
	case <-handler.executionStart:
		t.Log("Workflow execution started")
	case <-time.After(10 * time.Second):
		t.Log("No execution start message received")
	}

	// Wait for execution to complete or timeout
	select {
	case <-handler.executionEnd:
		t.Log("Workflow execution completed")
	case <-time.After(60 * time.Second):
		t.Log("Workflow execution timed out")
	}
}

// workflowMessageHandler workflow message handler
type workflowMessageHandler struct {
	executionStart chan struct{}
	executionEnd   chan struct{}
	progress       chan int
}

func (h *workflowMessageHandler) WriteMessage(messageType int, data []byte) error {
	// Here we can handle messages sent to ComfyUI
	return nil
}

func (h *workflowMessageHandler) ReadMessage() (messageType int, p []byte, err error) {
	// Here we should read messages from ComfyUI
	// For testing, we return a simple message
	return 1, []byte(`{"type": "execution_start"}`), nil
}

func (h *workflowMessageHandler) Close() error {
	return nil
}

func (h *workflowMessageHandler) Name() string {
	return "test-workflow-handler"
}
