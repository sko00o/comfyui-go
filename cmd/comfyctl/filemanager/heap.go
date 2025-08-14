package filemanager

import "time"

type FileEntry struct {
	path    string // full path
	size    int64
	modTime time.Time
	index   int // index in the heap
}

type FileInfoHeap []*FileEntry

func (h FileInfoHeap) Len() int { return len(h) }
func (h FileInfoHeap) Less(i, j int) bool {
	if h[i].modTime.Equal(h[j].modTime) {
		return h[i].size > h[j].size
	}
	return h[i].modTime.Before(h[j].modTime)
} // Min-Heap based on mod time
func (h FileInfoHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *FileInfoHeap) Push(x interface{}) {
	if x == nil {
		return
	}
	item, ok := x.(*FileEntry)
	if !ok {
		return
	}
	n := len(*h)
	item.index = n
	*h = append(*h, item)
}

func (h *FileInfoHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*h = old[0 : n-1]
	return item
}
