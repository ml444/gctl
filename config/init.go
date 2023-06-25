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

	"github.com/ml444/gctl/util"
	"gopkg.in/yaml.v3"

	log "github.com/ml444/glog"
	"github.com/spf13/viper"
)

const defaultTemplatesName = "gctl-templates"

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

	ProtoCentralRepoPath = viper.GetString(KeyProtoCentralRepoPath)
	ThirdPartyProtoPath = viper.GetStringSlice(KeyThirdPartyProtoPath)
	if ProtoCentralRepoPath != "" {
		ThirdPartyProtoPath = append(ThirdPartyProtoPath, ProtoCentralRepoPath)
	}
	TargetRootPath = viper.GetString(KeyTargetRootPath)
	TmplRootDir = viper.GetString(KeyTemplateRootDir)
	if TmplRootDir == "" {
		cmd := exec.Command("bash", "-c", "cd "+TargetRootPath+" && git clone https://github.com/ml444/gctl-templates.git")
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
		TmplRootDir = filepath.Join(TargetRootPath, defaultTemplatesName, "separation_templates")
		ThirdPartyProtoPath = append(ThirdPartyProtoPath, filepath.Join(TargetRootPath, defaultTemplatesName, "protofiles"))
		//fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=/your/path'", KeyTemplateRootDir))
		//return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyTemplateRootDir))
	} else {
		if strings.Contains(TmplRootDir, defaultTemplatesName) {
			sList := strings.Split(TmplRootDir, defaultTemplatesName)
			baseDir := strings.TrimSuffix(sList[0], string(os.PathSeparator))
			ThirdPartyProtoPath = append(ThirdPartyProtoPath, filepath.Join(baseDir, defaultTemplatesName, "protofiles"))
		}
	}

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
	if GoModulePrefix == "" {
		fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=your_repository_host'", KeyModulePrefix))
		return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyModulePrefix))
	}
	OnceFiles = viper.GetStringSlice(KeyOnceFiles)
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
	fmt.Printf("%#v\n", conf)
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
