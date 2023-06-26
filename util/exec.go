package util

import (
	"bytes"
	log "github.com/ml444/glog"
	"os/exec"
)

func CmdExec(cmdStr string) {
	cmd := exec.Command("bash", "-c", cmdStr)
	log.Infof("exec: %s", cmd.String())
	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf
	err := cmd.Run()
	if err != nil {
		log.Infof("Err: %s ", err.Error())
		log.Info("Stdout: ", outBuf.String())
		log.Info("Stderr: ", errBuf.String())
		return
	}
	log.Infof(" fmt files: %s", outBuf.String())
}
