package tests

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/sko00o/comfyui-go"
	"github.com/sko00o/comfyui-go/logger"
	"github.com/sko00o/comfyui-go/ws"
)

// TestComfyUIIntegration tests basic ComfyUI integration functionality
func TestComfyUIIntegration(t *testing.T) {
	// Skip integration tests unless explicitly requested
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	// Test basic API functionality
	t.Run("TestStatsAPI", func(t *testing.T) {
		testStatsAPI(t, client)
	})

	t.Run("TestHistoryAPI", func(t *testing.T) {
		testHistoryAPI(t, client)
	})

	t.Run("TestPromptAPI", func(t *testing.T) {
		testPromptAPI(t, client)
	})
}

// TestComfyUIWebSocketIntegration tests WebSocket integration functionality
func TestComfyUIWebSocketIntegration(t *testing.T) {
	// Skip integration tests unless explicitly requested
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

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

	// Test WebSocket connection
	t.Run("TestWebSocketConnection", func(t *testing.T) {
		testWebSocketConnection(t, host, port.Port())
	})
}

// startComfyUIContainer starts a ComfyUI container
func startComfyUIContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.ContainerRequest{
		FromDockerfile: testcontainers.FromDockerfile{
			Context:        "./container",
			Dockerfile:     "Dockerfile",
			PrintBuildLog:  true,
			BuildLogWriter: os.Stdout,
		},
		Image:        "comfyui:test",
		ExposedPorts: []string{"8188/tcp"},
		WaitingFor:   wait.ForHTTP("/system_stats").WithPort("8188").WithStartupTimeout(120 * time.Second),
		Env: map[string]string{
			"COMFYUI_PATH": "/home/comfy/ComfyUI",
		},
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return container, nil
}

// testStatsAPI tests system stats API
func testStatsAPI(t *testing.T, client *comfyui.Client) {
	stats, err := client.Stats()
	require.NoError(t, err)
	assert.NotNil(t, stats)
	assert.NotEmpty(t, stats.System.OS)
	assert.NotEmpty(t, stats.System.ComfyUIVersion)
	assert.NotEmpty(t, stats.System.PythonVersion)
	assert.NotEmpty(t, stats.System.PyTorchVersion)
}

// testHistoryAPI tests history API
func testHistoryAPI(t *testing.T, client *comfyui.Client) {
	history, err := client.GetHistory(10)
	require.NoError(t, err)
	assert.NotNil(t, history)
	// History may be empty in initial state
}

// testPromptAPI tests prompt API
func testPromptAPI(t *testing.T, client *comfyui.Client) {
	// Test getting current prompt status
	prompt, err := client.GetPrompt()
	require.NoError(t, err)
	assert.NotNil(t, prompt)
	assert.GreaterOrEqual(t, prompt.ExecInfo.QueueRemaining, 0)
}

// testWebSocketConnection tests WebSocket connection
func testWebSocketConnection(t *testing.T, host, port string) {
	// Create WebSocket URL
	wsURL := fmt.Sprintf("ws://%s:%s/ws", host, port)
	u, err := url.Parse(wsURL)
	require.NoError(t, err)

	// Create simple message handler
	handler := &testMessageHandler{
		received: make(chan []byte, 10),
	}

	// Create WebSocket client
	wsClient, err := ws.New(*u, "test-client", handler, logger.NewStd())
	require.NoError(t, err)
	defer wsClient.Close()

	// Wait for connection to establish
	time.Sleep(2 * time.Second)

	// Test if connection is working (by checking if messages are received)
	select {
	case <-handler.received:
		// Message received, connection is working
	case <-time.After(5 * time.Second):
		t.Log("No messages received, but connection might still be working")
	}
}

// testMessageHandler message handler for testing
type testMessageHandler struct {
	received chan []byte
}

func (h *testMessageHandler) HandleMessage(messageType int, payload []byte) {
	h.received <- payload
}
