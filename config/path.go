package config

import (
	"path/filepath"
	"strings"
)

func GetTargetDir(serviceGroup string, relativeDir []string, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(GlobalConfig.TargetRootPath, GlobalConfig.GoModulePrefix))
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range relativeDir {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	return filepath.Join(elems...)
}

func JoinModulePrefixWithGroup(serviceGroup string) string {
	modulePrefix := GlobalConfig.GoModulePrefix
	if serviceGroup != "" {
		return filepath.Join(modulePrefix, serviceGroup)
	}
	return modulePrefix
}
