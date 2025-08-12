package graph

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CachedHTTPObjectInfoFetcher struct {
	HTTPObjectInfoFetcher
	cache      NodeInfos
	cacheMutex sync.RWMutex
}

func NewCachedHTTPObjectInfoFetcher(baseURL string) *CachedHTTPObjectInfoFetcher {
	return &CachedHTTPObjectInfoFetcher{
		HTTPObjectInfoFetcher: *NewHTTPObjectInfoFetcher(baseURL),
		cache:                 make(NodeInfos),
	}
}

func (f *CachedHTTPObjectInfoFetcher) FetchNodeInfo(nodeType string) (*NodeInfo, error) {
	f.cacheMutex.RLock()
	if info, exists := f.cache[nodeType]; exists {
		f.cacheMutex.RUnlock()
		return info, nil
	}
	f.cacheMutex.RUnlock()

	// 缓存中不存在，从接口获取
	info, err := f.HTTPObjectInfoFetcher.FetchNodeInfo(nodeType)
	if err != nil {
		return nil, err
	}

	// 存入缓存
	f.cacheMutex.Lock()
	f.cache[nodeType] = info
	f.cacheMutex.Unlock()

	return info, nil
}

// HTTPObjectInfoFetcher 实现从 HTTP 接口获取 object_info
type HTTPObjectInfoFetcher struct {
	BaseURL string
	client  *http.Client
}

func NewHTTPObjectInfoFetcher(baseURL string) *HTTPObjectInfoFetcher {
	return &HTTPObjectInfoFetcher{
		BaseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}
}

func (f *HTTPObjectInfoFetcher) FetchNodeInfo(nodeType string) (*NodeInfo, error) {
	url := fmt.Sprintf("%s/api/object_info/%s", f.BaseURL, nodeType)
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node info for %q: %v", nodeType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch node info for %q: status code %d", nodeType, resp.StatusCode)
	}

	/*
		resp body is like this:
		{
			"LoadImage": {
				"input": {
					...
				}
			}
		}
	*/
	var nodeInfos NodeInfos
	if err := json.NewDecoder(resp.Body).Decode(&nodeInfos); err != nil {
		return nil, fmt.Errorf("failed to decode node info for %q: %v", nodeType, err)
	}

	nodeInfo, ok := nodeInfos[nodeType]
	if !ok {
		return nil, fmt.Errorf("node info for %q not found", nodeType)
	}
	return nodeInfo, nil
}

// FileObjectInfoFetcher 实现从本地文件夹获取 object_info
type FileObjectInfoFetcher struct {
	FileDir string
	cache   NodeInfos
}

type NodeInfos map[string]*NodeInfo

func NewFileObjectInfoFetcher(fileDir string) (*FileObjectInfoFetcher, error) {
	// read all json files in the directory
	files, err := filepath.Glob(filepath.Join(fileDir, "*.json"))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch node info from %s: %v", fileDir, err)
	}

	cache := make(NodeInfos)
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to read file %s: %v", file, err)
		}

		var nodeInfos NodeInfos
		if err := json.Unmarshal(content, &nodeInfos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal node info from %s: %v", file, err)
		}

		for nodeType, nodeInfo := range nodeInfos {
			cache[nodeType] = nodeInfo
		}
	}

	return &FileObjectInfoFetcher{
		FileDir: fileDir,
		cache:   cache,
	}, nil
}

func (f *FileObjectInfoFetcher) FetchNodeInfo(nodeType string) (*NodeInfo, error) {
	v, ok := f.cache[nodeType]
	if !ok {
		return nil, fmt.Errorf("node info for %q not found", nodeType)
	}
	return v, nil
}
