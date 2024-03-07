package cmd

import (
	"errors"
	"path/filepath"
	"strings"

	"github.com/ml444/gctl/config"
)

func RequiredParams(name *string, args []string, projectGroup *string) error {
	if name == nil || *name == "" {
		if len(args) == 0 {
			return errors.New("proto name must be provided")
		}
		*name = strings.TrimSpace(args[0])
	}

	if strings.Contains(*name, "-") {
		return errors.New("prohibited use of '-'")
	}

	if projectGroup == nil || *projectGroup == "" {
		*projectGroup = config.GlobalConfig.DefaultSvcGroup
	}
	return nil
}
func getProtoName(protoPath string) string {
	_, fname := filepath.Split(protoPath)
	return strings.TrimSuffix(fname, ".proto")
}
