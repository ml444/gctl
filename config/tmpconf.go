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

var TmplFilesConf TemplateConfigFile

type TemplateConfigFile struct {
	Target struct {
		RelativeDir struct {
			Proto  []string `yaml:"proto"`
			Client []string `yaml:"client"`
			Server []string `yaml:"server"`
		} `yaml:"relativeDir"`
	} `yaml:"target"`
	Template struct {
		FilesFormatSuffix string `yaml:"filesFormatSuffix"`
		ProtoFilename     string `yaml:"protoFilename"`
		RelativeDir       struct {
			Proto  []string `yaml:"proto"`
			Client []string `yaml:"client"`
			Server []string `yaml:"server"`
		} `yaml:"relativeDir"`
	} `yaml:"template"`
}

func InitTmplFilesConf() error {
	var err error
	tmplConfPath := filepath.Join(TmplRootDir, "config.yaml")
	err = ReadYaml(tmplConfPath, &TmplFilesConf)
	if err != nil {
		return err
	}
	return nil
}
func GetTempProtoFilename() string {
	return TmplFilesConf.Template.ProtoFilename
}

func GetTempFilesFormatSuffix() string {
	if TmplFilesConf.Template.FilesFormatSuffix == "" {
		return ".tmpl"
	}
	return TmplFilesConf.Template.FilesFormatSuffix
}

func GetTempProtoAbsPath() string {
	var elems []string
	elems = append(elems, TmplRootDir)
	elems = append(elems, TmplFilesConf.Template.RelativeDir.Proto...)
	elems = append(elems, GetTempProtoFilename())
	return filepath.Join(elems...)
}

func GetTempClientAbsDir() string {
	var elems []string
	elems = append(elems, TmplRootDir)
	elems = append(elems, TmplFilesConf.Template.RelativeDir.Client...)
	return filepath.Join(elems...)
}
func GetTempServerAbsDir() string {
	var elems []string
	elems = append(elems, TmplRootDir)
	elems = append(elems, TmplFilesConf.Template.RelativeDir.Server...)
	return filepath.Join(elems...)
}

func GetTargetProtoAbsPath(serviceGroup, protoName string) string {
	if filepath.IsAbs(protoName) {
		return protoName
	}
	var elems []string
	var useCentralRepo bool
	if ProtoCentralRepoPath != "" {
		useCentralRepo = true
	}
	if useCentralRepo {
		elems = append(elems, ProtoCentralRepoPath)
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
		protoName = fmt.Sprintf("%s.proto", serviceName)
	}
	if !useCentralRepo {
		for _, el := range TmplFilesConf.Target.RelativeDir.Proto {
			elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
		}
	}

	elems = append(elems, protoName)
	return filepath.Join(elems...)
}
func GetTargetClientAbsDir(serviceGroup, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(TargetRootPath, GoModulePrefix))
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range TmplFilesConf.Target.RelativeDir.Client {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	return filepath.Join(elems...)
}
func GetTargetServerAbsDir(serviceGroup, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(TargetRootPath, GoModulePrefix))
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range TmplFilesConf.Target.RelativeDir.Server {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	return filepath.Join(elems...)
}
