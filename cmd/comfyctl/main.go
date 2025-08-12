package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/marsgopher/mahou"
	"github.com/marsgopher/mahou/cmd"
	"github.com/marsgopher/mahou/log"
	"github.com/spf13/cobra"

	sd "github.com/sko00o/comfyui-go/driver"
	comfyError "github.com/sko00o/comfyui-go/error"
	"github.com/sko00o/comfyui-go/graph"
	"github.com/sko00o/comfyui-go/iface"
	"github.com/sko00o/comfyui-go/logger"
)

func WrapHandleErr(err error) (errObj map[string]any, isOOM bool) {
	errObj = make(map[string]any)
	errObj["err"] = fmt.Sprintf("handle: %v", err)
	var cuiErr comfyError.ComfyUIError
	if errors.As(err, &cuiErr) {
		isOOM = cuiErr.IsOOM
		if len(cuiErr.Message) > 0 {
			errObj["comfyui_err"] = cuiErr.Message
		} else {
			errObj["comfyui_err"] = nil
		}
		errObj["nodes_time"] = cuiErr.NodesTime
	}
	return
}

type Config struct {
	SD              sd.Config `mapstructure:"sd"`
	WorkflowPattern []string  `mapstructure:"workflow_pattern"`
	NoRecord        bool      `mapstructure:"no_record"`
	RecordDir       string    `mapstructure:"record_dir"`
	ClientID        string    `mapstructure:"client_id"`

	EnableSupervisor bool `mapstructure:"enable_supervisor"`
}

func main() {
	NewCommand(cmd.RootCmd)
	mahou.Execute()
}

func NewCommand(cmd *cobra.Command) {
	flags := cmd.Flags()
	flags.StringP("sd.comfy_ui.endpoint", "E", "http://localhost:8188", "ComfyUI endpoint")
	flags.Float64("sd.ram_free_threshold", 0.1, "RAM free threshold")
	flags.Float64("sd.vram_free_threshold", 0.2, "VRAM free threshold")
	flags.Float64("sd.torch_vram_free_threshold", 0, "torch VRAM free threshold")
	flags.StringArrayP("workflow_pattern", "D", []string{"*.json"}, "workflow dir pattern")
	flags.BoolP("no_record", "N", false, "no record")
	flags.StringP("client_id", "I", "", "client id")
	cmd.RunE = mahou.RunEFunc(func(ctx context.Context, c mahou.ConfigUnmarshaler) error {
		var cfg Config
		if err := c.Unmarshal(&cfg); err != nil {
			return err
		}
		return run(ctx, cfg)
	})
}

func processWorkflowFile(ctx context.Context, file string, recordFile string, driver *sd.Driver, converter *graph.GraphConverter, clientID string) error {
	// Read workflow file
	workflowData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}

	var req sd.PromptRequest
	req.Request = sd.Request{}
	if strings.HasSuffix(file, "_api.json") {
		req.Prompt = workflowData
	} else if strings.HasSuffix(file, "_full.json") {
		if err := json.Unmarshal(workflowData, &req); err != nil {
			return fmt.Errorf("unmarshal: %w", err)
		}
	} else {
		// convert from workflow to prompt
		var rawMessage json.RawMessage = workflowData
		req.Prompt, err = converter.Convert(rawMessage)
		if err != nil {
			return fmt.Errorf("converting workflow: %w", err)
		}
		req.ExtraData = json.RawMessage(fmt.Sprintf(`{"extra_data":{"workflow":%s}}`, rawMessage))
	}

	if err := driver.KeepSystemHealthy(ctx); err != nil {
		return fmt.Errorf("keeping system healthy: %w", err)
	}

	log.Infof("start processing %s", file)
	res, err := processRetryOnOOM(ctx, driver, uuid.New().String(), clientID, req)
	if err != nil {
		return fmt.Errorf("processing workflow: %w", err)
	}
	if res.NodeErrors != nil && string(res.NodeErrors) != "{}" {
		return fmt.Errorf("workflow node errors: %s", res.NodeErrors)
	} else {
		log.Debugf("response: %+v", res)

	}
	log.Infof("Successfully processed %s", file)

	// save record
	if err := os.WriteFile(recordFile, []byte(""), 0644); err != nil {
		return fmt.Errorf("writing record file: %w", err)
	}
	log.Debugf("record file: %s", recordFile)
	return nil
}

type Logger struct {
	*log.Logger
}

func (l *Logger) With(kv ...any) logger.LoggerExtend {
	return l
}

func run(ctx context.Context, cfg Config) error {
	if len(cfg.WorkflowPattern) == 0 {
		log.Fatal("workflow_pattern is required")
	}

	if cfg.SD.BaseDir == "" {
		cfg.SD.BaseDir = os.TempDir()
	}
	driver, err := sd.New(ctx, cfg.SD, sd.WithLogger(&Logger{Logger: log.Get()}))
	if err != nil {
		log.Fatalf("creating driver: %v", err)
	}

	var allFiles []string
	for _, pattern := range cfg.WorkflowPattern {
		// walk all json files in target dir
		files, err := filepath.Glob(pattern)
		if err != nil {
			log.Fatalf("finding workflow files with pattern %s: %v", pattern, err)
		}
		allFiles = append(allFiles, files...)
	}
	if len(allFiles) == 0 {
		log.Warnf("no files found matching the workflow patterns")
		return nil
	}

	// setup comfyui base url for object_info fetch
	fetcher := graph.NewCachedHTTPObjectInfoFetcher(cfg.SD.ComfyUI.Endpoint)
	converter := graph.NewGraphConverter(fetcher)

	recordDir := cfg.RecordDir
	if recordDir == "" {
		recordDir = filepath.Join(os.TempDir(), "comfyctl-record")
	}
	if err := os.MkdirAll(recordDir, 0755); err != nil {
		log.Fatalf("creating record dir: %v", err)
	}

	for _, file := range allFiles {
		// check record file
		base64Filename := base64.StdEncoding.EncodeToString([]byte(filepath.Base(file)))
		// escape /
		base64Filename = strings.ReplaceAll(base64Filename, "/", "_")
		recordFile := filepath.Join(recordDir, base64Filename)
		if _, err := os.Stat(recordFile); err == nil && !cfg.NoRecord {
			log.Infof("skipping %s because it has been processed", file)
			continue
		}

		// exit when error happened
		if err := processWorkflowFile(ctx, file, recordFile, driver, converter, cfg.ClientID); err != nil {
			log.Fatalf("processing file %s: %v", file, err)
		}
	}
	return nil
}

func process(driver *sd.Driver, taskID, clientID string, req sd.PromptRequest) (*sd.Response, error) {
	progressChan := make(chan iface.ProgressInfo, 1)
	defer close(progressChan)
	go func() {
		for progress := range progressChan {
			log.Infof("task %s progress is %d%% on node #%s", taskID, progress.PercentNum, progress.NodeID)
		}
	}()
	if clientID == "" {
		clientID = taskID
	}
	return driver.HandlePrompt(req.Request, taskID, clientID, "", progressChan)
}

func processRetryOnOOM(ctx context.Context, driver *sd.Driver, taskID, clientID string, req sd.PromptRequest) (*sd.Response, error) {
	maxRetries := 2
	var res *sd.Response
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		res, err = process(driver, taskID, clientID, req)
		if err != nil {
			if _, isOOM := WrapHandleErr(err); isOOM && attempt < maxRetries {
				if attempt == 0 {
					// First retry without reboot
					log.Infof("task %s is OOM (attempt %d/%d), retrying without reboot", taskID, attempt+1, maxRetries+1)
				} else {
					// Second retry with reboot
					log.Infof("task %s is OOM (attempt %d/%d), waiting for reboot", taskID, attempt+1, maxRetries+1)
					if err := driver.WaitingForReboot(ctx); err != nil {
						return res, fmt.Errorf("keeping system healthy: %w", err)
					}
				}
				// retry
				continue
			}
		}
		// If no error or non-OOM error, break out of the loop
		break
	}
	return res, err
}
