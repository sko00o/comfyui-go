package filemanager

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sko00o/comfyui-go/logger"
)

// ProgressReader wraps an io.Reader to provide progress logging
type ProgressReader struct {
	reader     io.Reader
	logger     logger.Logger
	filename   string
	totalBytes int64
	lastLog    time.Time
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.totalBytes += int64(n)

	// Log progress every 5 seconds to avoid spam
	now := time.Now()
	if pr.logger != nil && now.Sub(pr.lastLog) > 5*time.Second {
		pr.logger.Infof("downloading %s: %d bytes downloaded so far...", pr.filename, pr.totalBytes)
		pr.lastLog = now
	}

	return n, err
}

type IFileManager interface {
	SyncFile(filename string, load func() (io.ReadCloser, error)) error
}

type FileManager struct {
	dir          string
	maxBytes     int64
	currentBytes int64
	files        FileInfoHeap

	fileMutexes map[string]*sync.Mutex
	mu          sync.Mutex
	logger      logger.Logger

	maxRetries int
	retryDelay time.Duration
}

func (fm *FileManager) SyncFile(filename string, load func() (io.ReadCloser, error)) error {
	// Get or create a mutex for this specific file
	fm.mu.Lock()
	mutex, ok := fm.fileMutexes[filename]
	if !ok {
		mutex = &sync.Mutex{}
		fm.fileMutexes[filename] = mutex
	}
	fm.mu.Unlock()

	// Lock the file-specific mutex
	mutex.Lock()
	defer mutex.Unlock()

	fullname := filepath.Join(fm.dir, filename)

	// check if file exist
	if _, err := os.Stat(fullname); err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("stat %s: %w", fullname, err)
		}
	} else {
		if fm.logger != nil {
			fm.logger.Infof("file %s already exists, skipping sync", filename)
		}
		return nil
	}

	// Log start of sync
	if fm.logger != nil {
		fm.logger.Infof("starting sync for file: %s", filename)
	}

	// Add file to heap with retries
	return fm.addFile(fullname, load)
}

type Option func(*FileManager)

func WithLogger(logger logger.Logger) Option {
	return func(fm *FileManager) {
		fm.logger = logger
	}
}

func WithMaxRetries(maxRetries int) Option {
	return func(fm *FileManager) {
		fm.maxRetries = maxRetries
	}
}

func WithRetryDelay(retryDelay time.Duration) Option {
	return func(fm *FileManager) {
		fm.retryDelay = retryDelay
	}
}

func NewFileManager(dir string, maxMB int64, opts ...Option) (*FileManager, error) {
	fm := &FileManager{
		dir:         dir,
		maxBytes:    maxMB * 1024 * 1024,
		files:       make(FileInfoHeap, 0),
		fileMutexes: make(map[string]*sync.Mutex),
		maxRetries:  3,
		retryDelay:  time.Second * 2,
	}
	for _, opt := range opts {
		opt(fm)
	}
	heap.Init(&fm.files)
	err := fm.scanDirectory()
	if err != nil {
		return nil, err
	}
	return fm, nil
}

func (fm *FileManager) scanDirectory() error {
	if err := os.MkdirAll(fm.dir, 0755); err != nil {
		return err
	}
	err := filepath.Walk(fm.dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			entry := &FileEntry{
				path:    filepath.Join(fm.dir, path),
				size:    info.Size(),
				modTime: info.ModTime(),
			}
			heap.Push(&fm.files, entry)
			fm.currentBytes += info.Size()
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to walk through directory %s: %w", fm.dir, err)
	}

	return nil
}

// addFile adds a new file and controls the total directory size to not exceed the maximum limit.
// It accepts a load function to retry loading data in case of failures.
func (fm *FileManager) addFile(fullname string, load func() (io.ReadCloser, error)) error {
	tempPath := fullname + ".tmp"
	_ = os.MkdirAll(filepath.Dir(tempPath), 0755)

	// use helper function to handle file read and write and retry logic
	writtenBytes, copyErr := fm.copyWithRetry(tempPath, load)
	if copyErr != nil {
		return copyErr
	}

	// check if need delete oldest file
	for fm.currentBytes+writtenBytes > fm.maxBytes && fm.currentBytes > 0 {
		if delErr := fm.deleteOldestFile(); delErr != nil {
			os.Remove(tempPath)
			return delErr
		}
	}

	finalPath := fullname
	err := os.Rename(tempPath, finalPath)
	if err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename temporary file to final file: %w", err)
	}

	fm.currentBytes += writtenBytes

	// push new file to heap
	fileInfo, _ := os.Stat(finalPath)
	newEntry := &FileEntry{
		path:    finalPath,
		size:    fileInfo.Size(),
		modTime: fileInfo.ModTime(),
	}
	heap.Push(&fm.files, newEntry)

	return nil
}

func (fm *FileManager) tryCopy(tempPath string, load func() (io.ReadCloser, error)) (int64, error) {
	startTime := time.Now()

	reader, err := load()
	if err != nil {
		return 0, fmt.Errorf("load reader: %w", err)
	}
	defer reader.Close()

	tempFile, err := os.Create(tempPath)
	if err != nil {
		return 0, fmt.Errorf("failed to create temporary file: %w", err)
	}
	defer func() {
		if tempFile != nil {
			_ = tempFile.Close()
		}
	}()

	// Use a progress reader to track and log progress
	progressReader := &ProgressReader{
		reader:   reader,
		logger:   fm.logger,
		filename: filepath.Base(tempPath),
	}

	writtenBytes, err := io.Copy(tempFile, progressReader)
	if err != nil {
		_ = tempFile.Close()
		tempFile = nil
		os.Remove(tempPath)
		return 0, fmt.Errorf("failed to write data to temporary file: %w", err)
	}

	// Log completion with speed
	duration := time.Since(startTime)
	if fm.logger != nil && writtenBytes > 0 {
		speed := float64(writtenBytes) / duration.Seconds() / 1024 / 1024 // MB/s
		fm.logger.Infof("copied %d bytes in %v (%.2f MB/s) for %s",
			writtenBytes, duration, speed, filepath.Base(tempPath))
	}

	return writtenBytes, nil
}

func (fm *FileManager) copyWithRetry(tempPath string, load func() (io.ReadCloser, error)) (int64, error) {
	var writtenBytes int64
	var copyErr error
	retries := 0
	filename := filepath.Base(tempPath)

	for {
		if fm.logger != nil && retries > 0 {
			fm.logger.Infof("retrying download for %s (attempt %d/%d)", filename, retries+1, fm.maxRetries+1)
		}

		writtenBytes, copyErr = fm.tryCopy(tempPath, load)
		if copyErr == nil {
			break
		}

		retries++
		if retries >= fm.maxRetries {
			return 0, fmt.Errorf("io.Copy failed after %d retries: %v", fm.maxRetries, copyErr)
		}

		if fm.logger != nil {
			fm.logger.Warnf("failed to copy %s (retry %d/%d): %v, waiting %v before retry",
				filename, retries, fm.maxRetries, copyErr, fm.retryDelay)
		}

		time.Sleep(fm.retryDelay)
	}

	return writtenBytes, nil
}

func (fm *FileManager) deleteOldestFile() error {
	if fm.files.Len() == 0 {
		return fmt.Errorf("no file available to delete")
	}

	oldest := heap.Pop(&fm.files).(*FileEntry)
	if fm.logger != nil {
		fm.logger.Debugf("delete oldest file: %s", oldest.path)
	}
	if delError := os.Remove(oldest.path); delError != nil {
		// ignore file not exist
		if !os.IsNotExist(delError) {
			return delError
		}
	}
	fm.currentBytes -= oldest.size
	return nil
}
