package fakeshell

import (
	"os"
	"syscall"
	"time"
)

//FileInfoWrapper is a wrapper for os.FileInfo, which adds a fake Sys return value so that sh doesn't panic
type FileInfoWrapper struct {
	wrapped os.FileInfo
}

func NewFileInfoWrapper(wrapped os.FileInfo) *FileInfoWrapper {
	return &FileInfoWrapper{wrapped: wrapped}
}

func (f *FileInfoWrapper) Name() string {
	return f.wrapped.Name()
}

func (f *FileInfoWrapper) Size() int64 {
	return f.wrapped.Size()
}

func (f *FileInfoWrapper) Mode() os.FileMode {
	return f.wrapped.Mode()
}

func (f *FileInfoWrapper) ModTime() time.Time {
	return f.wrapped.ModTime()
}

func (f *FileInfoWrapper) IsDir() bool {
	return f.wrapped.IsDir()
}

func (f *FileInfoWrapper) Sys() any {
	return &syscall.Stat_t{
		Uid: uint32(os.Getuid()),
		Gid: uint32(os.Getgid()),
	}
}
