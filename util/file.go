package util

import (
	"fmt"
	"os"
	"path/filepath"
)

func OpenFile(fPath string) (*os.File, error) {
	dirPath := filepath.Dir(fPath)
	_, err := os.Stat(dirPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0775)
		if err != nil {
			return nil, err
		}
	}
	return os.OpenFile(fPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
}

func IsFileExist(name string) bool {
	fileInfo, err := os.Stat(name)
	if err != nil {
		fmt.Printf("err: %v", err)
		return os.IsExist(err)
	}
	if fileInfo != nil && fileInfo.IsDir() {
		fmt.Printf("This path '%v' is not a file path.", name)
		return false
	}
	return true
}
