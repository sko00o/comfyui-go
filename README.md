# ComfyUI-Go

Go based client for ComfyUI.

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/sko00o/comfyui-go"
)

func main() {
    // Create client
    client, err := comfyui.New(comfyui.Config{
        Endpoint: "http://localhost:8188",
        Timeout:  30 * time.Second,
    })
    if err != nil {
        panic(err)
    }

    // Get system stats
    stats, err := client.Stats()
    if err != nil {
        panic(err)
    }
    fmt.Printf("ComfyUI Version: %s\n", stats.System.ComfyUIVersion)
}
```

## References

- Routes: https://docs.comfy.org/development/comfyui-server/comms_routes
- Messages: https://docs.comfy.org/development/comfyui-server/comms_messages
