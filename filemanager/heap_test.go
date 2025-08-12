package filemanager

import (
	"container/heap"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFileInfoHeap_BasicOperations(t *testing.T) {
	h := make(FileInfoHeap, 0)
	heap.Init(&h)

	now := time.Now()
	files := []FileEntry{
		{path: "file1.png", size: 100, modTime: now.Add(-time.Hour)},     // -1h
		{path: "file2.png", size: 300, modTime: now.Add(-time.Hour * 3)}, // -3h
		{path: "file3.png", size: 200, modTime: now.Add(-time.Hour * 2)}, // -2h
	}

	for _, f := range files {
		heap.Push(&h, &f)
	}

	assert.Equal(t, 3, h.Len(), "Heap should contain 3 items")

	expected := []string{"file2.png", "file3.png", "file1.png"}
	for i, expectedFilename := range expected {
		file := heap.Pop(&h).(*FileEntry)
		assert.Equal(t, expectedFilename, file.path, "File %d should be %s", i+1, expectedFilename)
	}

	assert.Equal(t, 0, h.Len(), "Heap should be empty")
}

func TestFileInfoHeap_SameTimestamp(t *testing.T) {
	h := make(FileInfoHeap, 0)
	heap.Init(&h)

	sameTime := time.Now()
	filesWithSameTime := []FileEntry{
		{path: "same1.png", size: 100, modTime: sameTime},
		{path: "same2.png", size: 200, modTime: sameTime},
		{path: "same3.png", size: 300, modTime: sameTime},
	}

	for _, f := range filesWithSameTime {
		heap.Push(&h, &f)
	}

	expectedSame := []string{"same3.png", "same2.png", "same1.png"}
	for i, expectedFilename := range expectedSame {
		file := heap.Pop(&h).(*FileEntry)
		assert.Equal(t, expectedFilename, file.path, "File %d should be %s", i+1, expectedFilename)
	}
}

func TestFileInfoHeap_UpdateEntry(t *testing.T) {
	h := make(FileInfoHeap, 0)
	heap.Init(&h)

	now := time.Now()
	file1 := &FileEntry{path: "update1.png", size: 100, modTime: now}
	file2 := &FileEntry{path: "update2.png", size: 200, modTime: now.Add(time.Hour)}

	heap.Push(&h, file1)
	heap.Push(&h, file2)

	file1.modTime = now.Add(time.Hour * 2)
	heap.Init(&h) // Re-heapify after update

	first := heap.Pop(&h).(*FileEntry)
	second := heap.Pop(&h).(*FileEntry)
	assert.Equal(t, "update2.png", first.path)
	assert.Equal(t, "update1.png", second.path)
}
