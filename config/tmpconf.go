package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ml444/gkit/log"
	"gopkg.in/yaml.v3"
)

const (
	ProtoFileSuffix = ".proto"
	ServiceNameVar  = "{SERVICE_NAME}"
)

var TmplFilesConf TemplateConfig

func ReadYaml(fp string, conf interface{}) error {
	yamlFile, err := os.ReadFile(fp)
	if err != nil {
		log.Error(err)
		return err
	}
	err = yaml.Unmarshal(yamlFile, conf)
	if err != nil {
		log.Error(err)
		return err
	}
	//log.Debugf("%+v\n", conf)
	return nil
}

func InitTmplFilesConf() error {
	var err error
	if GlobalConfig.TemplatesBaseDir == "" {
		GlobalConfig.TemplatesBaseDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	if GlobalConfig.TemplatesConf == nil || GlobalConfig.TemplatesConf.Template.ProtoFilename == "" {
		tmplConfPath := filepath.Join(GlobalConfig.TemplatesBaseDir, "config.yaml")
		err = ReadYaml(tmplConfPath, &TmplFilesConf)
		if err != nil {
			return err
		}
		GlobalConfig.TemplatesConf = &TmplFilesConf
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
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, TmplFilesConf.Template.RelativeDir.Proto...)
	elems = append(elems, GetTempProtoFilename())
	return filepath.Join(elems...)
}

func GetTempClientAbsDir() string {
	var elems []string
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, TmplFilesConf.Template.RelativeDir.Client...)
	return filepath.Join(elems...)
}
func GetTempServerAbsDir() string {
	var elems []string
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, TmplFilesConf.Template.RelativeDir.Server...)
	return filepath.Join(elems...)
}

func GetGoPackage(serviceGroup, protoName string) string {
	var elems []string
	if GlobalConfig.GoModulePrefix != "" {
		elems = append(elems, GlobalConfig.GoModulePrefix)
	}
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	dir, name := filepath.Split(protoName)
	serviceName := strings.TrimSuffix(name, ProtoFileSuffix)
	if dir == "" {
		for _, el := range TmplFilesConf.Target.RelativeDir.Client {
			elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
		}
	} else if dir != "." {
		elems = append(elems, strings.Split(dir, "/")...)
		elems = append(elems, serviceName)
	} else {
		elems = append(elems, serviceName)
	}
	for i := 0; i < len(elems); i++ {
		if elems[i] == "" {
			elems = append(elems[:i], elems[i+1:]...)
			i--
		}
		if elems[i] == "." {
			elems = append(elems[:i], elems[i+1:]...)
			i--
		}
	}
	return strings.Join(elems, "/")
}

func GetTargetProtoAbsPath(serviceGroup, protoPath string) string {
	if filepath.IsAbs(protoPath) {
		return protoPath
	}
	var elems []string
	var useCentralRepo bool
	if GlobalConfig.ProtoCentralRepoPath != "" {
		useCentralRepo = true
	}
	if useCentralRepo {
		elems = append(elems, GlobalConfig.ProtoCentralRepoPath)
	} else {
		if GlobalConfig.TargetBaseDir == "" {
			GlobalConfig.TargetBaseDir, _ = os.Getwd()
		}
		elems = append(elems, GlobalConfig.TargetBaseDir)
		if GlobalConfig.GoModulePrefix != "" {
			elems = append(elems, GlobalConfig.GoModulePrefix)
		}
	}
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	if !strings.HasSuffix(protoPath, ProtoFileSuffix) {
		protoPath = fmt.Sprintf("%s.proto", protoPath)
		//serviceName = strings.TrimSuffix(serviceName, ProtoFileSuffix)
	}
	if len(strings.Split(protoPath, "/")) > 1 {
		protoPath = strings.TrimLeft(protoPath, "./")
		elems = append(elems, strings.Split(protoPath, "/")...)

	} else if !useCentralRepo {
		serviceName := strings.TrimSuffix(protoPath, ProtoFileSuffix)
		for _, el := range TmplFilesConf.Target.RelativeDir.Proto {
			elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
		}
		elems = append(elems, protoPath)
	}

	return filepath.Join(elems...)
}
func GetTargetClientAbsDir0(packagePath string) string {
	if GlobalConfig.TargetBaseDir == "" {
		GlobalConfig.TargetBaseDir, _ = os.Getwd()
	}
	if runtime.GOOS == "windows" {
		packagePath = strings.ReplaceAll(packagePath, "/", "\\")
	}
	return filepath.Join(GlobalConfig.TargetBaseDir, packagePath)
}
func GetTargetClientAbsDir(serviceGroup, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(GlobalConfig.TargetBaseDir, GlobalConfig.GoModulePrefix))
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
	elems = append(elems, filepath.Join(GlobalConfig.TargetBaseDir, GlobalConfig.GoModulePrefix))
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range TmplFilesConf.Target.RelativeDir.Server {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	return filepath.Join(elems...)
}

func GetTargetDir(serviceGroup string, relativeDir []string, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(GlobalConfig.TargetBaseDir, GlobalConfig.GoModulePrefix))
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
