package driver

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/sko00o/comfyui-go"
	"github.com/sko00o/comfyui-go/helper"
	"github.com/sko00o/comfyui-go/iface"
	"github.com/sko00o/comfyui-go/node"
)

type AnyInput struct {
	// optional, if not set, we will use the default bucket
	Bucket string `json:"bucket"`

	// Files override the nodeIDs with FieldName (recommended)
	Files []string `json:"files"`

	// sync to which directory under local ComfyUI
	SyncTo string `json:"sync_to"`
}

type AnyOutput struct {
	// output process will trigger on these nodes
	NodeIDs []string `json:"node_ids"`

	// We will replace the output path
	DirPath string `json:"dir_path"`

	// the default field_name is "images"
	FieldName string `json:"field_name"`
}

func (o *AnyOutput) newWorkflow(obj map[string]any, enableMetadata bool) (map[string]any, []string) {
	nodeIDs := []string{}
	for _, nodeID := range o.NodeIDs {
		// has node
		if v, ok := obj[nodeID]; ok {
			if vv, ok := v.(map[string]any); ok {
				// has inputs
				if v, ok := vv["inputs"]; ok {
					if vv, ok := v.(map[string]any); ok {
						// has field name
						fieldName := "images"
						if o.FieldName != "" {
							fieldName = o.FieldName
						}
						if v, ok := vv[fieldName]; ok {
							var preNode node.PreNode
							if vv, ok := v.([]any); ok {
								if len(vv) == 2 {
									if id, ok := vv[0].(string); ok {
										preNode.ID = id
										if argc, ok := vv[1].(float64); ok {
											preNode.Argc = int(argc)
										}
									}
								}
							}
							obj[nodeID] = node.SaveImageWebsocket{
								Images:         preNode,
								EnableMetadata: enableMetadata,
							}.Build()
							nodeIDs = append(nodeIDs, nodeID)
						}
					}
				}
			}
		}
	}
	return obj, nodeIDs
}

type Request struct {
	Workflow json.RawMessage `json:"workflow"`
	// compatible with ComfyUI prompt API,
	// if workflow is empty, we will use prompt as workflow
	Prompt json.RawMessage `json:"prompt"`

	// extra data will be write into pnginfo
	ExtraData json.RawMessage `json:"extra_data"`

	// will not write metadata to pnginfo
	DisableMetadata bool `json:"disable_metadata"`

	// the default bucket for input and output
	Bucket string `json:"bucket"`

	// multiple output config
	Outputs []AnyOutput `json:"outputs"`

	// Enable node replace
	EnableNodeReplace bool `json:"enable_node_replace"`

	Inputs []AnyInput `json:"inputs"`
}

type Response struct {
	comfyui.QueuePromptResp
	Outputs   []NodeOutputDetail `json:"outputs"`
	StartTime time.Time          `json:"start_time"`
	EndTime   time.Time          `json:"end_time"`

	// record time for sync complete
	SyncDuration time.Duration `json:"sync_duration"`

	// count time for each node
	NodesTime map[string]time.Duration `json:"nodes_time"`

	Hostname string `json:"hostname"`
}

func (d *Driver) HandlePrompt(req Request, taskID, clientID, newPromptID string, progressChan chan<- iface.ProgressInfo) (*Response, error) {
	start := time.Now()

	// workflow to obj
	workflowRaw := req.Workflow
	if len(workflowRaw) == 0 {
		workflowRaw = req.Prompt
	}
	workflowObj := make(map[string]any)
	if err := json.Unmarshal(workflowRaw, &workflowObj); err != nil {
		return nil, fmt.Errorf("unmarshal workflow: %w", err)
	}

	syncDuration, err := d.syncInputFiles(req)
	if err != nil {
		return nil, fmt.Errorf("sync input files: %w", err)
	}

	newData := func(clientID string) (data map[string]any, totalNodes int, isTriggerNodeID map[string]string) {
		isTriggerNodeID = make(map[string]string)

		// no replace output node on default
		if req.EnableNodeReplace {
			for _, output := range req.Outputs {
				prompt, triggerNodeIDs := output.newWorkflow(workflowObj, !req.DisableMetadata)
				for _, nodeID := range triggerNodeIDs {
					isTriggerNodeID[nodeID] = output.DirPath
				}
				workflowObj = prompt
			}
		} else {
			for _, output := range req.Outputs {
				for _, nodeID := range output.NodeIDs {
					isTriggerNodeID[nodeID] = output.DirPath
				}
			}
		}

		out := map[string]any{
			"prompt":    workflowObj,
			"client_id": clientID,
		}
		if req.ExtraData != nil {
			out["extra_data"] = req.ExtraData
		}
		return out, len(workflowObj), isTriggerNodeID
	}
	result, err := d.CommonGenerate(newData, req.Bucket, "", taskID, clientID, newPromptID, progressChan)
	if err != nil {
		return nil, err
	}

	outputs := []NodeOutputDetail{}
	for _, detail := range result.NodeOutput {
		outputs = append(outputs, *detail)
	}
	end := time.Now()
	return &Response{
		QueuePromptResp: result.QPResp,
		Outputs:         outputs,
		StartTime:       start,
		EndTime:         end,
		SyncDuration:    syncDuration,
		NodesTime:       result.NodesTime,
		Hostname:        helper.Hostname(),
	}, nil
}

// syncInputFiles synchronizes input files based on the request configuration.
// It calculates and returns the time taken for the synchronization.
func (d *Driver) syncInputFiles(req Request) (time.Duration, error) {
	cost := time.Duration(0)
	if req.Bucket != "" {
		for _, input := range req.Inputs {
			inputBucket := req.Bucket
			if input.Bucket != "" {
				inputBucket = input.Bucket
			}
			syncTo := SubDirInput
			if input.SyncTo != "" {
				syncTo = filepath.Clean(input.SyncTo)
			}
			files := input.Files
			start := time.Now()
			if err := d.prepareInputFiles(inputBucket, syncTo, files...); err != nil {
				return 0, fmt.Errorf("prepare input files for input %+v: %w", input, err)
			}
			cost += time.Since(start)
		}
	}
	return cost, nil
}

func (d *Driver) prepareInputFiles(bucket, syncTo string, files ...string) error {
	fm, ok := d.fManagerMap[syncTo]
	if !ok {
		return fmt.Errorf("sync_to %q is not support", syncTo)
	}

	if len(files) > 0 {
		d.Logger.Infof("preparing %d input files from bucket %s to %s", len(files), bucket, syncTo)
	}

	for i, name := range files {
		d.Logger.Infof("syncing file %d/%d: %s", i+1, len(files), name)
		if err := fm.SyncFile(name, func() (io.ReadCloser, error) {
			return d.Handler.Bucket(bucket).Open(name)
		}); err != nil {
			return fmt.Errorf("sync %s to %s: %w", name, syncTo, err)
		}
	}

	if len(files) > 0 {
		d.Logger.Infof("completed syncing %d input files to %s", len(files), syncTo)
	}

	return nil
}
