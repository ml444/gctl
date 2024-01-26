package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/ml444/gctl/config"
)

func RequiredParams(protoPath *string, args []string, serviceGroup *string) error {
	if protoPath == nil || *protoPath == "" {
		if len(args) == 0 {
			return errors.New("proto name must be provided")
		}
		*protoPath = strings.TrimSpace(args[0])
	}

	if strings.Contains(*protoPath, "-") {
		return errors.New("prohibited use of '-'")
	}

	if serviceGroup == nil {
		*serviceGroup = config.GlobalConfig.DefaultSvcGroup
	}

	return nil
}
func getProtoName(protoPath string) string {
	_, fname := filepath.Split(protoPath)
	return strings.TrimSuffix(fname, ".proto")
}
