package config

import (
	"errors"
	"fmt"
	"github.com/ml444/gctl/util"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/ml444/glog"
	"github.com/spf13/viper"
)

var (
	DbDSN                  string
	EnableAssignPort       bool
	SvcPortInterval        int
	SvcGroupInitPortMap    map[string]interface{}
	EnableAssignErrcode    bool
	SvcErrcodeInterval     int
	SvcGroupInitErrcodeMap map[string]interface{}

	DefaultSvcGroup string
	TmplRootDir     string
	TmplConfigFile  *TemplateConfigFile

	GoModulePrefix       string
	TargetRootPath       string
	OnceFiles            []string
	ThirdPartyProtoPath  []string
	ProtoCentralRepoPath string
)

func InitGlobalVar() error {
	var err error

	// read config file
	confPath := filepath.Join(GetHomeDir(), ".gctl_config.yaml")
	if util.IsFileExist(confPath) {
		viper.SetConfigFile(filepath.Join(GetHomeDir(), ".gctl_config.yaml"))
		err = viper.ReadInConfig()
		if err != nil {
			println(fmt.Sprintf("Warnning: %s", err))
		}
	}

	DbDSN = viper.GetString(KeyDbDSN)
	EnableAssignPort = viper.GetBool(KeyEnableAssignPort)
	EnableAssignErrcode = viper.GetBool(KeyEnableAssignErrcode)
	SvcGroupInitPortMap = viper.GetStringMap(KeySvcGroupInitPortMap)
	SvcGroupInitErrcodeMap = viper.GetStringMap(KeySvcGroupInitErrcodeMap)
	SvcPortInterval = viper.GetInt(KeySvcPortInterval)
	SvcErrcodeInterval = viper.GetInt(KeySvcErrcodeInterval)

	DefaultSvcGroup = viper.GetString(KeyDefaultServiceGroup)
	TmplRootDir = viper.GetString(KeyTemplateRootDir)
	TmplConfigFile = new(TemplateConfigFile)
	tmplConfPath := filepath.Join(TmplRootDir, "config.yaml")
	err = ReadYaml(tmplConfPath, TmplConfigFile)
	if err != nil {
		log.Error(err)
		return err
	}
	if TmplConfigFile == nil {
		return errors.New("this template repository is missing a configuration file")
	}
	GoModulePrefix = viper.GetString(KeyModulePrefix)
	TargetRootPath = viper.GetString(KeyTargetRootPath)
	OnceFiles = viper.GetStringSlice(KeyOnceFiles)
	ThirdPartyProtoPath = viper.GetStringSlice(KeyThirdPartyProtoPath)
	ProtoCentralRepoPath = viper.GetString(KeyProtoCentralRepoPath)
	return nil
}

func Validate() error {
	if TmplRootDir == "" {
		fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=/your/path'", KeyTemplateRootDir))
		return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyTemplateRootDir))
	}
	if viper.GetString(KeyModulePrefix) == "" {
		fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=your_repository_host'", KeyModulePrefix))
		return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyModulePrefix))
	}

	if TmplConfigFile == nil {
		return errors.New("template repository configuration file not found")
	}
	return nil
}

func PrintImportantVars() {
	fmt.Printf("===> GCTL_%s=%s\n", KeyTemplateRootDir, TmplRootDir)
	fmt.Printf("===> GCTL_%s=%s\n", KeyTargetRootPath, TargetRootPath)
	fmt.Printf("===> GCTL_%s=%s\n", KeyModulePrefix, GoModulePrefix)
	fmt.Printf("===> GCTL_%s=%#v\n", KeyOnceFiles, OnceFiles)
	fmt.Printf("===> GCTL_%s=%s\n", KeyProtoCentralRepoPath, ProtoCentralRepoPath)
	fmt.Printf("===> GCTL_%s=%s\n", KeyThirdPartyProtoPath, ThirdPartyProtoPath)
	fmt.Printf("===> GCTL_%s=%s\n", KeyDefaultServiceGroup, DefaultSvcGroup)
	fmt.Printf("===> GCTL_%s=%s\n", KeyDbDSN, DbDSN)
	fmt.Printf("===> GCTL_%s=%t\n", KeyEnableAssignPort, EnableAssignPort)
	fmt.Printf("===> GCTL_%s=%t\n", KeyEnableAssignErrcode, EnableAssignErrcode)
	fmt.Printf("===> GCTL_%s=%d\n", KeySvcPortInterval, SvcPortInterval)
	fmt.Printf("===> GCTL_%s=%d\n", KeySvcErrcodeInterval, SvcErrcodeInterval)
	fmt.Printf("===> GCTL_%s=%#v\n", KeySvcGroupInitPortMap, SvcGroupInitPortMap)
	fmt.Printf("===> GCTL_%s=%#v\n", KeySvcGroupInitErrcodeMap, SvcGroupInitErrcodeMap)

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
