package storage

import "io"

type Storage interface {
	GetFile(path string) (io.Reader, error)
	PutFile(dstPath string, filePath string) error
	HasFile(path string) bool
	ListFiles(path string) []string
}
