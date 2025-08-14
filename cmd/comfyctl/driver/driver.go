package driver

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"github.com/google/uuid"

	"github.com/marsgopher/fileop"
	fs "github.com/marsgopher/fileop/simplefs"

	comfyui "github.com/sko00o/comfyui-go"
	"github.com/sko00o/comfyui-go/iface"
	"github.com/sko00o/comfyui-go/logger"
	"github.com/sko00o/comfyui-go/session"
	"github.com/sko00o/comfyui-go/supervisor"

	"github.com/sko00o/comfyui-go/cmd/comfyctl/filemanager"
)

const (
	SubDirInput = "input"
	SubDirLoras = "models/loras"
)

const defaultFilenameTmplStr = `{{ .PromptID }}_{{ .Index }}{{ .EXT }}`

var defaultFilenameTmpl = template.Must(template.New("filename").Parse(defaultFilenameTmplStr))

type Config struct {
	ComfyUI comfyui.Config `mapstructure:"comfy_ui"`
	FS      fs.Config      `mapstructure:"fs"`

	// local base dir for ComfyUI
	BaseDir     string       `mapstructure:"base_dir"`
	DirManagers []DirManager `mapstructure:"dir_managers"`

	MaxTimeout time.Duration `mapstructure:"max_timeout"`

	RetryTimes int `mapstructure:"retry_times"`

	// RAMFreeThreshold is the threshold of free RAM usage
	RAMFreeThreshold float64 `mapstructure:"ram_free_threshold"`
	// VRAMFreeThreshold is the threshold of free VRAM usage
	VRAMFreeThreshold float64 `mapstructure:"vram_free_threshold"`
	// TorchVRAMFreeThreshold is the threshold of free torch VRAM usage
	TorchVRAMFreeThreshold float64 `mapstructure:"torch_vram_free_threshold"`

	DisableHealthCheck bool `mapstructure:"disable_health_check"`
}

type DirManager struct {
	SubDir string `mapstructure:"sub_dir"`
	MaxMB  int64  `mapstructure:"max_mb"`
}

type Option func(d *Driver)

func WithLogger(l logger.LoggerExtend) Option {
	return func(d *Driver) {
		d.Logger = l
	}
}

func New(ctx context.Context, c Config, opts ...Option) (*Driver, error) {
	// some default fix
	if c.FS.Mode == "" {
		c.FS.Mode = "disk"
	}
	if c.MaxTimeout == 0 {
		c.MaxTimeout = time.Minute * 30
	}

	fsHandler, err := fs.New(c.FS)
	if err != nil {
		return nil, fmt.Errorf("new fs: %w", err)
	}

	d := &Driver{
		Config:  c,
		Handler: fsHandler,
		Logger:  logger.NewStd(),
	}
	for _, opt := range opts {
		opt(d)
	}
	cli, err := comfyui.New(c.ComfyUI, comfyui.WithLogger(d.Logger))
	if err != nil {
		return nil, fmt.Errorf("new comfyui cli: %w", err)
	}
	d.Client = cli
	d.Supervisor = supervisor.NewSupervisor(cli, supervisor.WithLogger(d.Logger))

	if !c.DisableHealthCheck {
		if err := d.Supervisor.WaitingForSystemAlive(ctx); err != nil {
			return nil, fmt.Errorf("waiting for system alive: %w", err)
		}
	}

	fMangerMap := make(map[string]filemanager.IFileManager)
	for _, dm := range c.DirManagers {
		mDir := filepath.Join(c.BaseDir, dm.SubDir)
		maxMB := int64(10240)
		if dm.MaxMB > 0 {
			maxMB = dm.MaxMB
		}
		fm, err := filemanager.NewFileManager(mDir, maxMB, filemanager.WithLogger(d.Logger))
		if err != nil {
			return nil, fmt.Errorf("NewFileManager: %w", err)
		}
		fMangerMap[dm.SubDir] = fm
	}
	// add necessary
	mustHavePath := []string{
		SubDirInput,
		SubDirLoras,
	}
	for _, dir := range mustHavePath {
		if _, ok := fMangerMap[dir]; !ok {
			mDir := filepath.Join(c.BaseDir, dir)
			maxMB := int64(10240)
			fm, err := filemanager.NewFileManager(mDir, maxMB, filemanager.WithLogger(d.Logger))
			if err != nil {
				return nil, fmt.Errorf("NewFileManager: %w", err)
			}
			fMangerMap[dir] = fm
		}
	}
	d.fManagerMap = fMangerMap

	return d, nil
}

type Driver struct {
	*comfyui.Client
	iface.Supervisor

	Handler fileop.FileSystemSimpleBucket
	Config
	fManagerMap map[string]filemanager.IFileManager

	Logger logger.LoggerExtend
}

func (d *Driver) Stop() {
	d.Logger.Infof("driver shutdown...")

	d.Logger.Infof("driver exit")
}

type NewDataFunc func(clientID string) (data map[string]any, totalNodes int, isTriggerNodeID map[string]string)

func (d *Driver) CommonGenerate(newData NewDataFunc, bucket string, tmplStr, taskID, clientID, newPromptID string, progressChan chan<- iface.ProgressInfo) (*DriverSessionResult, error) {
	tmpl := defaultFilenameTmpl
	if tmplStr != "" {
		var err error
		tmpl, err = template.New("name_tmpl").Parse(tmplStr)
		if err != nil {
			return nil, fmt.Errorf("filename tmpl parse: %w", err)
		}
	}
	return d.commonGenerate(newData, bucket, tmpl, taskID, clientID, newPromptID, progressChan)
}

type NodeOutputDetail struct {
	NodeID string   `json:"node_id"`
	Dir    string   `json:"dir_path"`
	Files  []string `json:"files"`
	Texts  []string `json:"texts,omitempty"`
}

type NodeOutput map[string]*NodeOutputDetail

type SaveAdapter struct {
	fs fileop.FileSystemSimpleBucket
}

func (sa *SaveAdapter) Save(srcReader io.Reader, destPath string, contentType string) error {
	return sa.fs.PutStreamWithContentType(srcReader, destPath, contentType)
}

type DriverSessionResult struct {
	NodeOutput NodeOutput
	QPResp     comfyui.QueuePromptResp
	NodesTime  map[string]time.Duration
}

func (d *Driver) NewSession(
	taskID, clientID, promptID, bucket string,
	isTriggerNodeID map[string]string,
	nameMapCh map[string]chan string,
	textMapCh map[string]chan string,
	nameTmpl *template.Template,
	totalNodes int,
	progressChan chan<- iface.ProgressInfo,
) *session.Session {
	handler := d.Handler
	if bucket != "" {
		handler = handler.Bucket(bucket)
	}
	sess := session.New(
		taskID,
		clientID,
		promptID,
		isTriggerNodeID,
		nameMapCh,
		textMapCh,
		nameTmpl,
		totalNodes,
		progressChan,
		d.RetryTimes,
		d.Logger.With(
			"task_id", taskID,
			"client_id", clientID,
			"prompt_id", promptID,
		),
		d.Client,
		&SaveAdapter{fs: handler},
	)

	return sess
}

func (d *Driver) commonGenerate(
	newData NewDataFunc,
	bucket string,
	nameTmpl *template.Template,
	taskID, clientID, newPromptID string,
	progressChan chan<- iface.ProgressInfo,
) (result *DriverSessionResult, finalErr error) {
	result = &DriverSessionResult{}

	nameMapCh := make(map[string]chan string)
	textMapCh := make(map[string]chan string)
	// each dir, has many filenames
	nodeOutput := make(NodeOutput)
	defer func() {
		result.NodeOutput = nodeOutput
	}()

	wg := new(sync.WaitGroup)
	defer wg.Wait()

	if clientID == "" {
		clientID = uuid.New().String()
	}

	data, totalNodes, isTriggerNodeID := newData(clientID)
	for id, dir := range isTriggerNodeID {
		nameMapCh[id] = make(chan string, 1)
		textMapCh[id] = make(chan string, 1)
		nodeDetail := &NodeOutputDetail{
			NodeID: id,
			Dir:    dir,
			Files:  make([]string, 0),
			Texts:  make([]string, 0),
		}
		nodeOutput[id] = nodeDetail
		wg.Add(1)
		go func(nameCh chan string, detail *NodeOutputDetail) {
			defer wg.Done()
			for name := range nameCh {
				detail.Files = append(detail.Files, name)
			}
		}(nameMapCh[id], nodeDetail)
		go func(StrCh chan string, detail *NodeOutputDetail) {
			defer wg.Done()
			for text := range StrCh {
				detail.Texts = append(detail.Texts, text)
			}
		}(textMapCh[id], nodeDetail)
	}

	sess := d.NewSession(
		taskID,
		clientID,
		newPromptID,
		bucket,
		isTriggerNodeID,
		nameMapCh,
		textMapCh,
		nameTmpl,
		totalNodes,
		progressChan,
	)
	defer func() {
		result.NodesTime = sess.NodesTime
	}()

	processWg, err := d.SimpleProcess(clientID, &session.WrapSession{Session: sess})
	if err != nil {
		return nil, fmt.Errorf("consume process: %w", err)
	}
	defer processWg.Wait()

	var promptID string
	defer func() {
		// wait for session complete
		resMap := sess.Wait(d.MaxTimeout)
		if promptID != "" {
			res := resMap[promptID]
			if finalErr == nil {
				finalErr = errors.Join(res.Errs...)
			}
			result.QPResp = res.QPResp
		}

		for _, ch := range nameMapCh {
			close(ch)
		}
	}()

	resp, err := d.Prompt(data)
	if err != nil {
		finalErr = fmt.Errorf("send prompt resp: %w", err)
		return
	}
	promptID = resp.PromptID
	sess.StoreResp(promptID, session.RespResult{
		QPResp:    *resp,
		ErrorChan: make(chan error, 1),
	})
	return
}
