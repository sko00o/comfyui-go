package filemanager

import "io"

type FakeFileManager struct {
}

func (ffm FakeFileManager) SyncFile(filename string, load func() (io.ReadCloser, error)) error {
	return nil
}
