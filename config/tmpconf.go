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
	GroupNameVar    = "{GROUP_NAME}"
	ServiceNameVar  = "{SERVICE_NAME}"
)

type TemplateConfig struct {
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
		ServiceFileName   string `yaml:"serviceFileName"`
		RelativeDir       struct {
			Project []string `yaml:"project"`
			Proto   []string `yaml:"proto"`
			Client  []string `yaml:"client"`
			Server  []string `yaml:"server"`
		} `yaml:"relativeDir"`
	} `yaml:"template"`
}

// var TmplFilesConf TemplateConfig

func InitTemplateCfg() (*TemplateConfig, error) {
	tcfg := &TemplateConfig{}
	err := tcfg.Init()
	if err != nil {
		return nil, err
	}
	return tcfg, nil
}

func GetTmplFilesConf() (*TemplateConfig, error) {
	if GlobalConfig.TemplatesConf == nil {
		return InitTemplateCfg()
	}
	return GlobalConfig.TemplatesConf, nil
}

func readYaml(fp string, conf interface{}) error {
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
	return nil
}

func (tmplCfg *TemplateConfig) Init() error {
	var err error
	if GlobalConfig.TemplatesBaseDir == "" {
		GlobalConfig.TemplatesBaseDir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	if GlobalConfig.TemplatesConf == nil || GlobalConfig.TemplatesConf.Template.ProtoFilename == "" {
		tmplConfPath := filepath.Join(GlobalConfig.TemplatesBaseDir, "config.yaml")
		err = readYaml(tmplConfPath, tmplCfg)
		if err != nil {
			return err
		}
	}
	return nil
}

func (tmplCfg *TemplateConfig) ProtoTmplName() string {
	return tmplCfg.Template.ProtoFilename
}

func (tmplCfg *TemplateConfig) TempFileExtSuffix() string {
	if tmplCfg.Template.FilesFormatSuffix == "" {
		return ".tmpl"
	}
	return tmplCfg.Template.FilesFormatSuffix
}

func (tmplCfg *TemplateConfig) ProtoTmplAbsPath() string {
	var elems []string
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, tmplCfg.Template.RelativeDir.Proto...)
	elems = append(elems, tmplCfg.ProtoTmplName())
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ProjectTmplAbsDir() string {
	var elems []string
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, tmplCfg.Template.RelativeDir.Project...)
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ClientTmplAbsDir() string {
	var elems []string
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, tmplCfg.Template.RelativeDir.Client...)
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ServerTmplAbsDir() string {
	var elems []string
	elems = append(elems, GlobalConfig.TemplatesBaseDir)
	elems = append(elems, tmplCfg.Template.RelativeDir.Server...)
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) JoinGoPackage(serviceGroup, protoName string) string {
	var elems []string
	if GlobalConfig.GoModulePrefix != "" {
		elems = append(elems, GlobalConfig.GoModulePrefix)
		if serviceGroup != "" {
			elems = append(elems, serviceGroup)
		}
	}
	dir, name := filepath.Split(protoName)
	serviceName := strings.TrimSuffix(name, ProtoFileSuffix)
	if dir == "" {
		for _, el := range tmplCfg.Target.RelativeDir.Client {
			el = strings.ReplaceAll(el, GroupNameVar, serviceGroup)
			el = strings.ReplaceAll(el, ServiceNameVar, serviceName)
			elems = append(elems, el)
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

func (tmplCfg *TemplateConfig) ProtoTargetAbsPath(serviceGroup, protoPath string) string {
	if filepath.IsAbs(protoPath) {
		return protoPath
	}
	log.Infof("svcGroup: %s, protoFile: %s\n", serviceGroup, protoPath)
	var elems []string
	var useCentralRepo bool
	if GlobalConfig.ProtoCentralRepoPath != "" {
		useCentralRepo = true
	}
	if useCentralRepo {
		elems = append(elems, GlobalConfig.ProtoCentralRepoPath)
		if serviceGroup != "" {
			elems = append(elems, serviceGroup)
		}
	} else {
		if GlobalConfig.TargetBaseDir == "" {
			GlobalConfig.TargetBaseDir, _ = os.Getwd()
		}
		elems = append(elems, GlobalConfig.TargetBaseDir)
		if GlobalConfig.GoModulePrefix != "" {
			elems = append(elems, GlobalConfig.GoModulePrefix)
		}
	}
	if !strings.HasSuffix(protoPath, ProtoFileSuffix) {
		protoPath = fmt.Sprintf("%s.proto", protoPath)
		// serviceName = strings.TrimSuffix(serviceName, ProtoFileSuffix)
	}
	if len(strings.Split(protoPath, "/")) > 1 {
		protoPath = strings.TrimLeft(protoPath, "./")
		elems = append(elems, strings.Split(protoPath, "/")...)

	} else if !useCentralRepo {
		serviceName := strings.TrimSuffix(protoPath, ProtoFileSuffix)
		for _, el := range tmplCfg.Target.RelativeDir.Proto {
			el = strings.ReplaceAll(el, GroupNameVar, serviceGroup)
			el = strings.ReplaceAll(el, ServiceNameVar, serviceName)
			elems = append(elems, el)
		}
	}

	elems = append(elems, protoPath)
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ClientTargetAbsDir0(packagePath string) string {
	if GlobalConfig.TargetBaseDir == "" {
		GlobalConfig.TargetBaseDir, _ = os.Getwd()
	}
	switch runtime.GOOS {
	case "windows":
		packagePath = strings.ReplaceAll(packagePath, "/", "\\")
	}
	return filepath.Join(GlobalConfig.TargetBaseDir, packagePath)
}

func (tmplCfg *TemplateConfig) ClientTargetAbsDir(serviceGroup, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(GlobalConfig.TargetBaseDir, GlobalConfig.GoModulePrefix))
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range tmplCfg.Target.RelativeDir.Client {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ClientRelativeDirs(group, serviceName string) []string {
	var elems []string
	for _, el := range tmplCfg.Target.RelativeDir.Client {
		el = strings.ReplaceAll(el, GroupNameVar, group)
		el = strings.ReplaceAll(el, ServiceNameVar, serviceName)
		elems = append(elems, el)
	}
	return elems
}

func (tmplCfg *TemplateConfig) ServerTargetAbsPath(serviceGroup, serviceName string) string {
	var elems []string
	elems = append(elems, filepath.Join(GlobalConfig.TargetBaseDir, GlobalConfig.GoModulePrefix))
	if GlobalConfig.GoModulePrefix != "" && serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	for _, el := range tmplCfg.Target.RelativeDir.Server {
		el = strings.ReplaceAll(el, GroupNameVar, serviceGroup)
		el = strings.ReplaceAll(el, ServiceNameVar, serviceName)
		elems = append(elems, el)
	}
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ServerRelativeDirs(group, serviceName string) []string {
	var elems []string
	for _, el := range tmplCfg.Target.RelativeDir.Server {
		el = strings.ReplaceAll(el, GroupNameVar, group)
		el = strings.ReplaceAll(el, ServiceNameVar, serviceName)
		elems = append(elems, el)
	}
	return elems
}

func (tmplCfg *TemplateConfig) ProjectTargetAbsDir(serviceGroup string, projectName string) string {
	var elems []string
	if GlobalConfig.TargetBaseDir == "" {
		GlobalConfig.TargetBaseDir, _ = os.Getwd()
	}
	elems = append(elems, GlobalConfig.TargetBaseDir, GlobalConfig.GoModulePrefix, serviceGroup)
	if !strings.HasSuffix(GlobalConfig.TargetBaseDir, "/"+projectName) {
		elems = append(elems, projectName)
	}
	return filepath.Join(elems...)
}

func (tmplCfg *TemplateConfig) ProcessFilePath(targetRootDir, tmplFilePath, rootTmplPath, tmplName string) string {
	fileName := tmplCfg.GetFileName(tmplName)
	parentPath := strings.TrimSuffix(strings.TrimPrefix(tmplFilePath, rootTmplPath), tmplName)
	targetFile := targetRootDir + parentPath + fileName
	return targetFile
}

func (tmplCfg *TemplateConfig) GetFileName(tmplName string) string {
	return strings.TrimSuffix(tmplName, tmplCfg.TempFileExtSuffix())
}
