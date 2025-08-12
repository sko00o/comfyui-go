package driver

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/sko00o/comfyui-go"
)

type PromptRequest struct {
	Request
	ClientID string `json:"client_id"`
}

func (d *Driver) HandlerPrompt(ctx *gin.Context) {
	var req PromptRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promptID := uuid.New().String()

	go func() {
		res, err := d.HandlePrompt(req.Request, promptID, req.ClientID, promptID, nil)
		if err != nil {
			d.Logger.Errorf("prompt %s: %v", promptID, err)
		}
		if res.NodeErrors != nil && string(res.NodeErrors) != "{}" {
			d.Logger.Errorf("prompt %s result: %s", promptID, res.NodeErrors)
			return
		}
		d.Logger.Infof("prompt %s finished: %+v", promptID, res)
	}()

	resp := comfyui.QueuePromptResp{
		PromptID: promptID,
	}
	ctx.JSON(http.StatusOK, resp)
}

func (d *Driver) HandleView(ctx *gin.Context) {
	name := ctx.Query("filename")
	subDir := ctx.Query("subfolder")

	// we will get it from fs
	filename := filepath.Join(d.BaseDir, subDir, name)

	// response with file content
	f, err := d.Handler.Open(filename)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("open file: %v", err),
		})
		return
	}
	if _, err := io.Copy(ctx.Writer, f); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("copy file: %v", err),
		})
	}
}
