package openapi

import (
	"os"
	"path/filepath"
	"strings"
)

// 文件处理
type fileHandle []string

func (f *fileHandle) load(dir string) {
	filePath, err := filepath.Abs(dir)
	if err != nil {
		return
	}
	if !isDir(dir) {
		return
	}
	list, err := os.ReadDir(filePath)
	if err != nil {
		return
	}
	for _, v := range list {
		newPath := filepath.Join(filePath, v.Name())
		if v.IsDir() {
			f.load(newPath)
		} else if ext := filepath.Ext(v.Name()); ext == ".go" {
			if !strings.HasSuffix(strings.TrimSuffix(v.Name(), ext), "_test") {
				*f = append(*f, newPath)
			}
		}
	}
	return
}
