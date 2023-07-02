package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"

	log "github.com/ml444/glog"
	"github.com/ml444/gutil/osx"
	"github.com/spf13/viper"
)

const defaultTemplatesName = "gctl-templates"
const gctlConfigFileName = ".gctl_config.yaml"

//var (
//	DbURI            string
//	EnableAssignPort bool
//	SvcPortInterval        int
//	SvcGroupInitPortMap    map[string]interface{}
//	EnableAssignErrcode    bool
//	SvcErrcodeInterval     int
//	SvcGroupInitErrcodeMap map[string]interface{}
//
//	DefaultSvcGroup string
//	TmplRootDir     string
//
//	GoModulePrefix       string
//	TargetRootPath       string
//	OnceFiles            []string
//	AllProtoPathList  []string
//	ProtoCentralRepoPath string
//)

var GlobalConfig Config

type Config struct {
	DbURI                  string
	EnableAssignPort       bool
	SvcPortInterval        int
	SvcGroupInitPortMap    map[string]interface{}
	EnableAssignErrcode    bool
	SvcErrcodeInterval     int
	SvcGroupInitErrcodeMap map[string]interface{}

	DefaultSvcGroup string
	TmplRootDir     string

	GoModulePrefix       string
	TargetRootPath       string
	OnceFiles            []string
	AllProtoPathList     []string
	ProtoCentralRepoPath string
}

func InitGlobalVar() error {
	var err error

	// read config file
	confPath := filepath.Join(GetHomeDir(), gctlConfigFileName)
	if osx.IsFileExist(confPath) {
		viper.SetConfigFile(filepath.Join(GetHomeDir(), gctlConfigFileName))
		err = viper.ReadInConfig()
		if err != nil {
			println(fmt.Sprintf("Warnning: %s", err))
		}
	}

	GlobalConfig.DbURI = viper.GetString(KeyDbDSN)
	GlobalConfig.EnableAssignPort = viper.GetBool(KeyEnableAssignPort)
	GlobalConfig.EnableAssignErrcode = viper.GetBool(KeyEnableAssignErrcode)
	GlobalConfig.SvcGroupInitPortMap = viper.GetStringMap(KeySvcGroupInitPortMap)
	GlobalConfig.SvcGroupInitErrcodeMap = viper.GetStringMap(KeySvcGroupInitErrcodeMap)
	GlobalConfig.SvcPortInterval = viper.GetInt(KeySvcPortInterval)
	GlobalConfig.SvcErrcodeInterval = viper.GetInt(KeySvcErrcodeInterval)
	GlobalConfig.DefaultSvcGroup = viper.GetString(KeyDefaultServiceGroup)

	GlobalConfig.ProtoCentralRepoPath = viper.GetString(KeyProtoCentralRepoPath)
	GlobalConfig.AllProtoPathList = viper.GetStringSlice(KeyThirdPartyProtoPath)
	if GlobalConfig.ProtoCentralRepoPath != "" {
		GlobalConfig.AllProtoPathList = append(GlobalConfig.AllProtoPathList, GlobalConfig.ProtoCentralRepoPath)
	}
	GlobalConfig.TargetRootPath = viper.GetString(KeyTargetRootPath)
	GlobalConfig.TmplRootDir = viper.GetString(KeyTemplateRootDir)
	if GlobalConfig.TmplRootDir == "" {
		cmd := exec.Command("bash", "-c", "cd "+GlobalConfig.TargetRootPath+" && git clone https://github.com/ml444/gctl-templates.git")
		log.Infof("exec: %s", cmd.String())
		var outBuf, errBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &errBuf
		err = cmd.Run()
		if err != nil {
			log.Infof("Err: %s \nStdout: %s \n Stderr: %s", err, outBuf.String(), errBuf.String())
			return err
		}
		log.Infof(" %s", errBuf.String())
		GlobalConfig.TmplRootDir = filepath.Join(GlobalConfig.TargetRootPath, defaultTemplatesName, "separation_templates")
		GlobalConfig.AllProtoPathList = append(GlobalConfig.AllProtoPathList, filepath.Join(GlobalConfig.TargetRootPath, defaultTemplatesName, "protofiles"))
		//fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=/your/path'", KeyTemplateRootDir))
		//return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyTemplateRootDir))
	} else {
		if strings.Contains(GlobalConfig.TmplRootDir, defaultTemplatesName) {
			sList := strings.Split(GlobalConfig.TmplRootDir, defaultTemplatesName)
			baseDir := strings.TrimSuffix(sList[0], string(os.PathSeparator))
			GlobalConfig.AllProtoPathList = append(GlobalConfig.AllProtoPathList, filepath.Join(baseDir, defaultTemplatesName, "protofiles"))
		}
	}

	err = InitTmplFilesConf()
	if err != nil {
		log.Errorf("err: %v", err)
		return err
	}

	GlobalConfig.GoModulePrefix = viper.GetString(KeyModulePrefix)
	if GlobalConfig.GoModulePrefix == "" {
		fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=your_repository_host'", KeyModulePrefix))
		return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyModulePrefix))
	}
	GlobalConfig.OnceFiles = viper.GetStringSlice(KeyOnceFiles)
	return nil
}

func PrintImportantVars() {
	fmt.Printf("===> GCTL_%s=%s\n", KeyTemplateRootDir, GlobalConfig.TmplRootDir)
	fmt.Printf("===> GCTL_%s=%s\n", KeyTargetRootPath, GlobalConfig.TargetRootPath)
	fmt.Printf("===> GCTL_%s=%s\n", KeyModulePrefix, GlobalConfig.GoModulePrefix)
	fmt.Printf("===> GCTL_%s=%#v\n", KeyOnceFiles, GlobalConfig.OnceFiles)
	fmt.Printf("===> GCTL_%s=%s\n", KeyProtoCentralRepoPath, GlobalConfig.ProtoCentralRepoPath)
	fmt.Printf("===> GCTL_%s=%s\n", KeyThirdPartyProtoPath, GlobalConfig.AllProtoPathList)
	fmt.Printf("===> GCTL_%s=%s\n", KeyDefaultServiceGroup, GlobalConfig.DefaultSvcGroup)
	fmt.Printf("===> GCTL_%s=%s\n", KeyDbDSN, GlobalConfig.DbURI)
	fmt.Printf("===> GCTL_%s=%t\n", KeyEnableAssignPort, GlobalConfig.EnableAssignPort)
	fmt.Printf("===> GCTL_%s=%t\n", KeyEnableAssignErrcode, GlobalConfig.EnableAssignErrcode)
	fmt.Printf("===> GCTL_%s=%d\n", KeySvcPortInterval, GlobalConfig.SvcPortInterval)
	fmt.Printf("===> GCTL_%s=%d\n", KeySvcErrcodeInterval, GlobalConfig.SvcErrcodeInterval)
	fmt.Printf("===> GCTL_%s=%#v\n", KeySvcGroupInitPortMap, GlobalConfig.SvcGroupInitPortMap)
	fmt.Printf("===> GCTL_%s=%#v\n", KeySvcGroupInitErrcodeMap, GlobalConfig.SvcGroupInitErrcodeMap)

}

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
	log.Debugf("%+v", conf)
	return nil
}
func GetHomeDir() string {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("USERPROFILE")
	default:
		return os.Getenv("HOME")
	}
}
