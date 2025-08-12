package driver

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"text/template"
	"time"

	"sync/atomic"

	"github.com/marsgopher/fileop"

	"github.com/sko00o/comfyui-go"
	comfyError "github.com/sko00o/comfyui-go/error"
	"github.com/sko00o/comfyui-go/iface"
	"github.com/sko00o/comfyui-go/logger"
	"github.com/sko00o/comfyui-go/ws/message"
)

var (
	ErrTimeout = errors.New("session timeout")
)

var SupportedOutputKeys = []string{
	// handle "SaveImage", "PreviewImage", "SaveAnimatedWEBP"
	"images",
	// handle "SaveGLB"
	"3d",
	// handle "SaveAudio"
	"audio",
	// handle "Image Comparer (rgthree)"
	"a_images", "b_images",
	// handle "VHS_VideoCombine"
	"gifs",
}

type SaveSession struct {
	ID     string
	NodeID string
}

type Session struct {
	*comfyui.Client
	Handler fileop.FileSystemSimpleBucket

	// nodeID -> outputDir
	IsTriggerNode map[string]string
	runningNode   *SaveSession

	// nodeID -> filenames
	NameMapCh map[string]chan string
	// nodeID -> texts
	TextMapCh map[string]chan string

	idx    *atomic.Uint32
	resMap sync.Map

	FilenameTmpl *template.Template
	ClientID     string
	PromptID     string // promptID rewrite
	TaskID       string

	Logger logger.Logger

	TotalNodes    int
	ExecutedNodes []string
	ProgressChan  chan<- iface.ProgressInfo

	RetryTimes int

	done chan struct{}

	lastNodeID        string
	lastNodeStartTime time.Time
	NodesTime         map[string]time.Duration
}

func (s *Session) StoreResp(promptID string, resp RespResult) {
	s.resMap.Store(promptID, resp)
}

func (s *Session) LoadResp(promptID string) (RespResult, bool) {
	v, ok := s.resMap.Load(promptID)
	if !ok {
		return RespResult{}, false
	}
	return v.(RespResult), true
}

func (s *Session) RangeResp(fn func(promptID string, resp RespResult) bool) {
	s.resMap.Range(func(k, v any) bool {
		return fn(k.(string), v.(RespResult))
	})
}

type SessionResult struct {
	QPResp comfyui.QueuePromptResp
	Errs   []error
}

func (s *Session) Wait(maxTimeout time.Duration) map[string]SessionResult {
	defer close(s.done)
	resMap := make(map[string]SessionResult)

	timeout := time.NewTimer(maxTimeout)
	defer timeout.Stop()

	s.RangeResp(func(promptID string, resp RespResult) bool {
		var errs []error
		errCh := resp.ErrorChan

	SLOOP:
		for {
			select {
			case err, ok := <-errCh:
				if !ok {
					break SLOOP
				}
				errs = append(errs, err)
			case <-timeout.C:
				errs = append(errs, ErrTimeout)
				break SLOOP
			}
		}

		resMap[promptID] = SessionResult{
			QPResp: resp.QPResp,
			Errs:   errs,
		}
		return true
	})
	return resMap
}

func (s *Session) handleResult(promptID string, err error, isFinal bool) {
	resMap, ok := s.LoadResp(promptID)
	if ok {
		resCh := resMap.ErrorChan
		if err != nil {
			resCh <- err
		}
		if isFinal {
			s.runningNode = nil
			close(resCh)
		}
	}
}

func (s *Session) handleTextMessage(msg []byte) {
	var m message.Message
	if err := json.Unmarshal(msg, &m); err != nil {
		s.Logger.Warnf("TXT message unmarshal: %v, skip", err)
		return
	}
	// skip nil data
	if m.Data == nil {
		return
	}

	switch m.Type {
	case message.Executing:
		if o, ok := m.Data.(*message.DataExecuting); ok {
			if s.lastNodeID != "" {
				s.NodesTime[s.lastNodeID] += time.Since(s.lastNodeStartTime)
			}
			s.lastNodeID = ""
			s.lastNodeStartTime = time.Now()

			if o.Node != nil {
				s.lastNodeID = *o.Node
				s.updateProgress(o.GetPromptID(), *o.Node)
			} else {
				// final state
				s.handleResult(m.Data.GetPromptID(), nil, true)
			}
		}
	case message.ExecutionSuccess:
		s.handleResult(m.Data.GetPromptID(), nil, false)
	case message.ExecutionError:
		// check if the error is OOM
		isOOM := false
		if o, ok := m.Data.(*message.DataExecutionError); ok {
			isOOM = o.IsOOM()
		}
		err := comfyError.ComfyUIError{
			Message:   json.RawMessage(msg),
			IsOOM:     isOOM,
			NodesTime: s.NodesTime,
		}
		s.handleResult(m.Data.GetPromptID(), err, false)
	case message.ExecutionInterrupted:
		err := comfyError.ComfyUIError{
			Message:   json.RawMessage(msg),
			NodesTime: s.NodesTime,
		}
		s.handleResult(m.Data.GetPromptID(), err, false)
	case message.Executed:
		if o, ok := m.Data.(*message.DataExecuted); ok {
			if o.Node != nil {
				dir, ok := s.IsTriggerNode[*o.Node]
				if ok {
					if content, ok := o.Output["text"]; ok {
						s.handleText(*o.Node, content)
					}
					if dir != "" {
						for _, name := range SupportedOutputKeys {
							if content, ok := o.Output[name]; ok {
								newContent, err := s.modifyFileInfo(*o.Node, o.PromptID, content)
								if err != nil {
									s.Logger.Errorf("handle fileinfo: %v", err)
								} else {
									o.Output[name] = newContent
								}
							}
						}
						m.Data = o
					}
				}
			}
		}
	case message.ExecutionCached:
		if o, ok := m.Data.(*message.DataExecution); ok {
			if len(o.Nodes) > 0 {
				for _, nodeID := range o.Nodes {
					s.NodesTime[nodeID] = time.Duration(0)
				}

				s.updateProgress(o.GetPromptID(), o.Nodes...)
			}
		}

	case message.Progress:
		// if o, ok := m.Data.(*message.DataProgress); ok {}
	}
}

func (s *Session) handleText(nodeID string, content json.RawMessage) {
	var texts []string
	if err := json.Unmarshal(content, &texts); err != nil {
		s.Logger.Warnf("texts unmarshal: %v, skip", err)
		return
	}
	for _, text := range texts {
		s.saveText(nodeID, text)
	}
}

func (s *Session) modifyFileInfo(nodeID, promptID string, content json.RawMessage) (json.RawMessage, error) {
	var files []message.FileInfo
	if err := json.Unmarshal(content, &files); err != nil {
		return nil, fmt.Errorf("unmarshal images: %w", err)
	}

	for idx, fi := range files {
		ni := NameInfo{
			ClientID: s.ClientID,
			PromptID: promptID,
			Index:    s.idx.Add(1),
			EXT:      filepath.Ext(fi.Filename),
			TaskID:   s.TaskID,
		}

		var realFilename string
		if err := s.GetView(fi, func(reader io.Reader, header http.Header) error {
			ni.ContentType = header.Get("Content-Type")
			name, saveErr := s.save(nodeID, ni, reader)
			if saveErr != nil {
				return fmt.Errorf("save: %w", saveErr)
			}
			realFilename = name
			return nil
		}); err != nil {
			s.handleResult(promptID, fmt.Errorf("get image: %w", err), false)
			continue
		}

		// update to real filename
		dirname := s.IsTriggerNode[nodeID]
		files[idx].Subfolder = dirname
		files[idx].Filename = realFilename
	}

	newContent, err := json.Marshal(files)
	if err != nil {
		return nil, fmt.Errorf("marshal images: %w", err)
	}
	return newContent, nil
}

func (s *Session) handleBinaryMessage(msg []byte) {
	var nodeID, promptID string
	{
		ss := s.runningNode
		if ss == nil {
			return
		}
		nodeID = ss.NodeID
		promptID = ss.ID
	}

	dir, ok := s.IsTriggerNode[nodeID]
	if !ok || dir == "" {
		return
	}

	var b message.BinaryMessage
	if err := b.UnmarshalBinary(msg); err != nil {
		s.Logger.Warnf("BIN message unmarshal: %v, skip", err)
		s.handleResult(promptID, fmt.Errorf("unmarshal binary: %w", err), false)
		return
	}

	if b.Type == message.PreviewImage {
		if img, ok := b.Data.(*message.DataImage); ok {
			s.Logger.Debugf("ws trigger save on node #%s", nodeID)
			ni := NameInfo{
				ClientID:    s.ClientID,
				PromptID:    promptID,
				Index:       s.idx.Add(1),
				EXT:         img.Type.Ext(),
				TaskID:      s.TaskID,
				ContentType: img.Type.ContentType(),
			}
			if _, err := s.save(nodeID, ni, bytes.NewReader(img.Blob)); err != nil {
				s.handleResult(promptID, fmt.Errorf("save: %w", err), false)
				return
			}
		}
	}
}

func (s *Session) updateProgress(promptID string, nodes ...string) {
	s.ExecutedNodes = append(s.ExecutedNodes, nodes...)
	currentNodeID := nodes[len(nodes)-1]
	s.runningNode = &SaveSession{
		ID:     promptID,
		NodeID: currentNodeID,
	}
	if s.ProgressChan != nil {
		progress := int(float64(len(s.ExecutedNodes)) / float64(s.TotalNodes) * 100)
		// progress will no larger than 99
		if progress >= 100 {
			progress = 99
		}
		s.ProgressChan <- iface.ProgressInfo{
			NodeID:     currentNodeID,
			PercentNum: progress,
		}
	}
}

type NameInfo struct {
	ClientID string
	PromptID string
	Index    uint32

	// ext has dot prefix, e.g.: ".png"
	EXT string

	TaskID string

	// if ContentType is empty, will use default
	ContentType string
}

func (s *Session) filename(ni NameInfo) (string, error) {
	var name bytes.Buffer
	if err := s.FilenameTmpl.Execute(&name, ni); err != nil {
		return "", fmt.Errorf("filename template: %w", err)
	}
	return name.String(), nil
}

func (s *Session) save(id string, ni NameInfo, rd io.Reader) (string, error) {
	name, err := s.filename(ni)
	if err != nil {
		return "", fmt.Errorf("filename template: %w", err)
	}
	s.Logger.Infof("trigger save on node #%s, Content-Type: %q", id, ni.ContentType)
	return name, s.saveAndProcess(id, name, rd, ni.ContentType)
}

func (s *Session) saveAndProcess(id, name string, rd io.Reader, contentType string) error {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "pimg-upload-*")
	if err != nil {
		return fmt.Errorf("create temp file: %w", err)
	}
	defer func() {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFile.Name())
	}()

	// Copy input stream to temp file
	if _, err := io.Copy(tmpFile, rd); err != nil {
		return fmt.Errorf("copy to temp file: %w", err)
	}
	s.Logger.Debugf("save %s to tmp file %s", name, tmpFile.Name())

	retryTimes := s.RetryTimes
	if retryTimes == 0 {
		retryTimes = 3
	}

	for i := 0; i < retryTimes; i++ {
		// Seek to start for each retry
		if _, err := tmpFile.Seek(0, 0); err != nil {
			return fmt.Errorf("seek temp file: %w", err)
		}

		if err := s.saveReader(id, name, tmpFile, contentType); err != nil {
			s.Logger.Warnf("save %s failed, retry %d: %v", name, i, err)
			continue
		}
		s.Logger.Debugf("save %s success", name)
		return nil
	}
	return fmt.Errorf("save: %s, retry %d times, failed", name, retryTimes)
}

func (s *Session) saveReader(id, name string, rd io.Reader, contentType string) (err error) {
	defer func() {
		if err == nil {
			// collect generated filename
			s.saveName(id, name)
		}
	}()
	dirname := s.IsTriggerNode[id]
	target := filepath.Join(dirname, name)
	err = s.Handler.PutStreamWithContentType(rd, target, contentType)
	return err
}

func (s *Session) saveName(id, name string) {
	if nameCh, ok := s.NameMapCh[id]; ok {
		nameCh <- name
	}
}

func (s *Session) saveText(id, text string) {
	if textCh, ok := s.TextMapCh[id]; ok {
		textCh <- text
	}
}
