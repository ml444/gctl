package config

import (
	"fmt"
	"path/filepath"
	"strings"
)

const (
	ProtoFileSuffix = ".proto"
	ServiceNameVar  = "{SERVICE_NAME}"
)

func GetProtoAbsPath(serviceGroup, protoName string) string {
	if filepath.IsAbs(protoName) {
		return protoName
	}
	var elems []string
	repoPath := ProtoCentralRepoPath
	if repoPath != "" {
		elems = append(elems, repoPath)
	} else {
		elems = append(elems, filepath.Join(TargetRootPath, GoModulePrefix))
	}
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	serviceName := protoName
	if strings.HasSuffix(protoName, ProtoFileSuffix) {
		serviceName = strings.TrimSuffix(serviceName, ProtoFileSuffix)
	} else {
		protoName = fmt.Sprintf("%s.proto", protoName)
	}
	if repoPath == "" {
		for _, el := range TmplConfigFile.Target.RelativeDir.Proto {
			elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
		}
	}

	elems = append(elems, protoName)
	return filepath.Join(elems...)
}

func GetTargetDir(serviceGroup string, relativeDir []string, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(TargetRootPath, GoModulePrefix))
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range relativeDir {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	return filepath.Join(elems...)
}

func JoinModulePrefixWithGroup(serviceGroup string) string {
	modulePrefix := GoModulePrefix
	if serviceGroup != "" {
		return filepath.Join(modulePrefix, serviceGroup)
	}
	return modulePrefix
}

func GetTemplateProtoDir() string {
	var elems []string
	elems = append(elems, TmplRootDir)
	elems = append(elems, TmplConfigFile.Template.RelativeDir.Proto...)
	//elems = append(elems, templateConfigFile.TemplateProtoFilename)
	return filepath.Join(elems...)
}
