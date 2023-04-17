package util

import (
	log "github.com/ml444/glog"
	"os"
	"path/filepath"
)

func OpenFile(fPath string) (*os.File, error) {
	dirPath := filepath.Dir(fPath)
	info, err := os.Stat(dirPath)
	if err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0775)
		if err != nil {
			log.Fatalf("%v \n", err)
		}
	}
	if info != nil && !info.IsDir() {
		log.Fatalf("This path isn't dir: %v \n", dirPath)
	}
	return os.OpenFile(fPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
}

func IsFileExist(name string) bool {
	fileInfo, err := os.Stat(name)
	//if err != nil {
	//	log.Errorf("err: %v", err)
	//}
	if fileInfo != nil && fileInfo.IsDir() {
		log.Fatalf("This path '%v' is not a file path.", name)
	}
	return err == nil || os.IsExist(err)
}
