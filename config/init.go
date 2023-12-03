package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/ml444/gkit/config"
	"github.com/ml444/gkit/config/yaml"
	"github.com/ml444/gkit/log"
	"github.com/ml444/gutil/osx"
)

const defaultTemplatesName = "gctl-templates"
const gctlConfigFileName = ".gctl_config.yaml"
const defaultTemplateProjectName = "separation_templates"

var GlobalConfig Config
var cfg *config.Config

type Config struct {
	DbURI                  string         `yaml:"DB_URI" env:"name=GCTL_DB_URI"`
	EnableAssignPort       bool           `yaml:"ENABLE_ASSIGN_PORT" env:"name=GCTL_ENABLE_ASSIGN_PORT;default=false"`
	EnableAssignErrcode    bool           `yaml:"ENABLE_ASSIGN_ERRCODE" env:"name=GCTL_ENABLE_ASSIGN_ERRCODE;default=false"`
	SvcErrcodeInterval     int            `yaml:"SVC_ERRCODE_INTERVAL" env:"name=GCTL_SVC_ERRCODE_INTERVAL;default=1000"`
	SvcPortInterval        int            `yaml:"SVC_PORT_INTERVAL" env:"name=GCTL_SVC_PORT_INTERVAL;default=10"`
	SvcGroupInitPortMap    map[string]int `yaml:"SVC_GROUP_INIT_PORT_MAP" env:"name=GCTL_SVC_GROUP_INIT_PORT_MAP"`
	SvcGroupInitErrcodeMap map[string]int `yaml:"SVC_GROUP_INIT_ERRCODE_MAP" env:"name=GCTL_SVC_GROUP_INIT_ERRCODE_MAP"`

	DefaultSvcGroup string `yaml:"DEFAULT_SERVICE_GROUP" env:"name=GCTL_DEFAULT_SERVICE_GROUP"`
	TmplRootDir     string `yaml:"TEMPLATES_ROOT_DIR" env:"name=GCTL_TEMPLATES_ROOT_DIR"`

	GoModulePrefix       string   `yaml:"MODULE_PREFIX" env:"name=GCTL_MODULE_PREFIX"`
	TargetRootPath       string   `yaml:"TARGET_ROOT_PATH" env:"name=GCTL_TARGET_ROOT_PATH"`
	OnceFiles            []string `yaml:"ONCE_FILES" env:"name=GCTL_ONCE_FILES;default=.gitignore,go.mod,.editorconfig,README.md,Dockerfile,Makefile"`
	AllProtoPathList     []string `yaml:"PROTO_PATHS" env:"name=GCTL_THIRD_PARTY_PROTO_PATH"`
	ProtoCentralRepoPath string   `yaml:"PROTO_CENTRAL_REPO_PATH" env:"name=GCTL_PROTO_CENTRAL_REPO_PATH"`

	//TmplFilesConf *TemplateConfigFile `yaml:"-"`
}

func InitConfig() error {
	var err error
	cfg, err = config.InitConfig(
		&GlobalConfig,
		config.WithFileLoader(yaml.NewLoader()),
		config.WithFilePath(filepath.Join(GetHomeDir(), gctlConfigFileName)),
	)
	if err != nil {
		return err
	}
	return nil
}

func InitGlobalVar() error {
	var err error
	err = InitConfig()
	if err != nil {
		return err
	}

	// read config file
	confPath := filepath.Join(GetHomeDir(), gctlConfigFileName)
	if osx.IsFileExist(confPath) {
		_, err = config.InitConfig(
			&GlobalConfig,
			config.WithFileLoader(yaml.NewLoader()),
			config.WithFilePath(filepath.Join(GetHomeDir(), gctlConfigFileName)),
		)
		if err != nil {
			return err
		}
	}

	//GlobalConfig.DbURI = viper.GetString(KeyDbDSN)
	//GlobalConfig.EnableAssignPort = viper.GetBool(KeyEnableAssignPort)
	//GlobalConfig.EnableAssignErrcode = viper.GetBool(KeyEnableAssignErrcode)
	//GlobalConfig.SvcGroupInitPortMap = viper.GetStringMap(KeySvcGroupInitPortMap)
	//GlobalConfig.SvcGroupInitErrcodeMap = viper.GetStringMap(KeySvcGroupInitErrcodeMap)
	//GlobalConfig.SvcPortInterval = viper.GetInt(KeySvcPortInterval)
	//GlobalConfig.SvcErrcodeInterval = viper.GetInt(KeySvcErrcodeInterval)
	//GlobalConfig.DefaultSvcGroup = viper.GetString(KeyDefaultServiceGroup)

	//GlobalConfig.ProtoCentralRepoPath = viper.GetString(KeyProtoCentralRepoPath)
	//GlobalConfig.AllProtoPathList = viper.GetStringSlice(KeyThirdPartyProtoPath)
	if GlobalConfig.ProtoCentralRepoPath != "" {
		GlobalConfig.AllProtoPathList = append(GlobalConfig.AllProtoPathList, GlobalConfig.ProtoCentralRepoPath)
	}
	//GlobalConfig.TargetRootPath = viper.GetString(KeyTargetRootPath)
	//GlobalConfig.TmplRootDir = viper.GetString(KeyTemplateRootDir)
	//if GlobalConfig.TmplRootDir == "" {
	//	cmd := exec.Command("bash", "-c", "cd "+GlobalConfig.TargetRootPath+" && git clone https://github.com/ml444/gctl-templates.git")
	//	log.Infof("exec: %s", cmd.String())
	//	var outBuf, errBuf bytes.Buffer
	//	cmd.Stdout = &outBuf
	//	cmd.Stderr = &errBuf
	//	err = cmd.Run()
	//	if err != nil {
	//		log.Infof("Err: %s \nStdout: %s \n Stderr: %s", err, outBuf.String(), errBuf.String())
	//		return err
	//	}
	//	log.Infof(" %s", errBuf.String())
	//	GlobalConfig.TmplRootDir = filepath.Join(GlobalConfig.TargetRootPath, defaultTemplatesName, defaultTemplateProjectName)
	//	GlobalConfig.AllProtoPathList = append(GlobalConfig.AllProtoPathList, filepath.Join(GlobalConfig.TargetRootPath, defaultTemplatesName, "protos"))
	//	//fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=/your/path'", KeyTemplateRootDir))
	//	//return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyTemplateRootDir))
	//} else {
	//	if strings.Contains(GlobalConfig.TmplRootDir, defaultTemplatesName) {
	//		sList := strings.Split(GlobalConfig.TmplRootDir, defaultTemplatesName)
	//		baseDir := strings.TrimSuffix(sList[0], string(os.PathSeparator))
	//		GlobalConfig.AllProtoPathList = append(GlobalConfig.AllProtoPathList, filepath.Join(baseDir, defaultTemplatesName, "protofiles"))
	//	}
	//}
	if GlobalConfig.TmplRootDir != "" {
		err = InitTmplFilesConf()
		if err != nil {
			log.Errorf("err: %v", err)
			return err
		}
	}

	//GlobalConfig.GoModulePrefix = viper.GetString(KeyModulePrefix)
	//if GlobalConfig.GoModulePrefix == "" {
	//	fmt.Println(fmt.Sprintf("err: must be set: 'export GCTL_%s=your_repository_host'", KeyModulePrefix))
	//	return errors.New(fmt.Sprintf("missing environment variable: GCTL_%s", KeyModulePrefix))
	//}
	//GlobalConfig.OnceFiles = viper.GetStringSlice(KeyOnceFiles)
	return nil
}

func PrintImportantVars() {
	_ = cfg.Walk(func(key string, value *config.Value) error {
		if value.EnvName() != "" {
			fmt.Printf("===> %s=%v\n", value.EnvName(), value.Value())
		} else {
			fmt.Printf("===> %s=%v\n", key, value)
		}
		return nil
	})
	//fmt.Printf("===> GCTL_TEMPLATES_ROOT_DIR=%s\n", GlobalConfig.TmplRootDir)
	//fmt.Printf("===> GCTL_TARGET_ROOT_PATH=%s\n", GlobalConfig.TargetRootPath)
	//fmt.Printf("===> GCTL_MODULE_PREFIX=%s\n", GlobalConfig.GoModulePrefix)
	//fmt.Printf("===> GCTL_ONCE_FILES=%#v\n", GlobalConfig.OnceFiles)
	//fmt.Printf("===> GCTL_PROTO_PATHS=%s\n", GlobalConfig.AllProtoPathList)
	//fmt.Printf("===> GCTL_PROTO_CENTRAL_REPO_PATH=%s\n", GlobalConfig.ProtoCentralRepoPath)
	//fmt.Printf("===> GCTL_DEFAULT_SERVICE_GROUP=%s\n", GlobalConfig.DefaultSvcGroup)
	//fmt.Printf("===> GCTL_DB_URI=%s\n", GlobalConfig.DbURI)
	//fmt.Printf("===> GCTL_ENABLE_ALLOC_PORT=%t\n", GlobalConfig.EnableAssignPort)
	//fmt.Printf("===> GCTL_ENABLE_ALLOC_ERRCODE=%t\n", GlobalConfig.EnableAssignErrcode)
	//fmt.Printf("===> GCTL_SVC_PORT_INTERVAL=%d\n", GlobalConfig.SvcPortInterval)
	//fmt.Printf("===> GCTL_SVC_ERRCODE_INTERVAL=%d\n", GlobalConfig.SvcErrcodeInterval)
	//fmt.Printf("===> GCTL_SVC_GROUP_INIT_PORT_MAP=%#v\n", GlobalConfig.SvcGroupInitPortMap)
	//fmt.Printf("===> GCTL_SVC_GROUP_INIT_ERRCODE_MAP=%#v\n", GlobalConfig.SvcGroupInitErrcodeMap)

}

func GetHomeDir() string {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("USERPROFILE")
	default:
		return os.Getenv("HOME")
	}
}
