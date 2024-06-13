package config

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/ml444/gkit/config"
	"github.com/ml444/gkit/config/yaml"
	"github.com/ml444/gkit/log"

	"github.com/ml444/gctl/util"
)

const gctlConfigFileName = ".gctl_config.yaml"

var (
	GlobalConfig = Config{}
	cfg          *config.Config
)

type Config struct {
	Debug bool `yaml:"Debug" env:"name=GCTL_DEBUG"`

	// @desc: Database URI
	// MySQL: mysql://username:password@tcp(ip:port)/database
	// Postgres: postgres://username:password@ip:port/database	or "user=astaxie password=astaxie dbname=test sslmode=disable"
	// SQLite: sqlite://username:password@ip:port/database
	DBURI string `yaml:"DbURI" env:"name=GCTL_DB_URI"`
	// @desc: The name of the csv file to store the service port and errcode when the database is not used.
	CsvFile string `yaml:"CsvFile" env:"name=GCTL_CSV_FILE;default=gctl_service_settings.csv"`

	EnableAssignPort       bool           `yaml:"EnableAssignPort" env:"name=GCTL_ENABLE_ASSIGN_PORT;default=false"`
	EnableAssignErrcode    bool           `yaml:"EnableAssignErrcode" env:"name=GCTL_ENABLE_ASSIGN_ERRCODE;default=false"`
	SvcErrcodeInterval     int            `yaml:"SvcErrcodeInterval" env:"name=GCTL_SVC_ERRCODE_INTERVAL;default=1000"`
	SvcPortInterval        int            `yaml:"SvcPortInterval" env:"name=GCTL_SVC_PORT_INTERVAL;default=10"`
	DefaultStartingPort    int            `yaml:"DefaultStartingPort" env:"name=GCTL_DEFAULT_STARTING_PORT;default=10000"`
	DefaultStartingErrcode int            `yaml:"DefaultStartingErrcode" env:"name=GCTL_DEFAULT_STARTING_ERRCODE;default=100000"`
	SvcGroupInitPortMap    map[string]int `yaml:"SvcGroupInitPortMap" env:"name=GCTL_SVC_GROUP_INIT_PORT_MAP"`
	SvcGroupInitErrcodeMap map[string]int `yaml:"SvcGroupInitErrcodeMap" env:"name=GCTL_SVC_GROUP_INIT_ERRCODE_MAP"`

	TargetBaseDir          string   `yaml:"TargetBaseDir" env:"name=GCTL_TARGET_BASE_DIR"`
	DefaultSvcGroup        string   `yaml:"DefaultServiceGroup" env:"name=GCTL_DEFAULT_SVC_GROUP"`
	GoModulePrefix         string   `yaml:"GoModulePrefix" env:"name=GCTL_MODULE_PREFIX"`
	OnceFiles              []string `yaml:"OnceFiles" env:"name=GCTL_ONCE_FILES;default=.gitignore,go.mod,.editorconfig,README.md,Dockerfile,Makefile"`
	ProtoPaths             []string `yaml:"ProtoPaths" env:"name=GCTL_PROTO_PATHS"`
	ProtoCentralRepoPath   string   `yaml:"ProtoCentralRepoPath" env:"name=GCTL_PROTO_CENTRAL_REPO_PATH"`
	SwaggerCentralRepoPath string   `yaml:"SwaggerCentralRepoPath" env:"name=GCTL_SWAGGER_CENTRAL_REPO_PATH"`

	TemplatesBaseDir string          `yaml:"TemplatesBaseDir" env:"name=GCTL_TEMPLATES_BASE_DIR"`
	TemplatesConf    *TemplateConfig `yaml:"TemplatesConf"`

	// @desc: usage of yaml
	// ProtoPlugins:
	//   - "--go_out=paths=source_relative:."
	//   - "--go-grpc_out=paths=source_relative:."
	//   - "--go-field_out=paths=source_relative,include_prefix=Model:."
	ProtoPlugins []string `yaml:"ProtoPlugins"`
}

func (c *Config) IsRelativePath() bool {
	return c.GoModulePrefix == ""
}

func InitConfig() error {
	var err error
	currentDir, _ := os.Getwd()
	cfgPath := filepath.Join(currentDir, gctlConfigFileName)
	if !IsFileExist(cfgPath) {
		cfgPath = filepath.Join(GetHomeDir(), gctlConfigFileName)
		if !IsFileExist(cfgPath) {
			cfg, err = config.InitConfig(&GlobalConfig)
			if err != nil {
				log.Warnf("init config err: %v", err)
				return nil
			}
			goto InitTemplates
		}
	}

	log.Infof("config file: %s \n", cfgPath)
	cfg, err = config.InitConfig(
		&GlobalConfig,
		config.WithFileLoader(yaml.NewLoader()),
		config.WithFilePath(cfgPath),
	)
	if err != nil {
		log.Warnf("init config err: %v", err)
		return nil
	}
InitTemplates:
	if tmpCfg := GlobalConfig.TemplatesConf; tmpCfg == nil || reflect.DeepEqual(tmpCfg, reflect.New(reflect.TypeOf(TemplateConfig{})).Interface()) {
		log.Info("template conf is nil")
		GlobalConfig.TemplatesConf, err = InitTemplateCfg()
		if err != nil {
			log.Errorf("err: %v", err)
			return err
		}
	}

	return nil
}

func IsFileExist(name string) bool {
	fileInfo, err := os.Stat(name)
	if err != nil {
		return os.IsExist(err)
	}
	if fileInfo != nil && fileInfo.IsDir() {
		fmt.Printf("This path '%v' is not a file path.", name)
		return false
	}
	return true
}

func InitGlobalVar() error {
	var err error
	err = InitConfig()
	if err != nil {
		return err
	}

	if GlobalConfig.ProtoCentralRepoPath != "" {
		GlobalConfig.ProtoPaths = append(GlobalConfig.ProtoPaths, GlobalConfig.ProtoCentralRepoPath)
	}
	if GlobalConfig.SwaggerCentralRepoPath != "" {
		if !util.IsDirExist(GlobalConfig.SwaggerCentralRepoPath) {
			os.MkdirAll(GlobalConfig.SwaggerCentralRepoPath, 0o664)
		}
	}

	return nil
}

func PrintImportantVars() {
	_ = cfg.Walk(func(key string, value *config.Value) error {
		if value.EnvName() != "" {
			fmt.Printf("===> %s=%v\n", value.EnvName(), value.Value())
		}
		return nil
	})
	// fmt.Printf("===> GCTL_TEMPLATES_ROOT_DIR=%s\n", GlobalConfig.TemplatesBaseDir)
	// fmt.Printf("===> GCTL_TARGET_ROOT_PATH=%s\n", GlobalConfig.TargetBaseDir)
	// fmt.Printf("===> GCTL_MODULE_PREFIX=%s\n", GlobalConfig.GoModulePrefix)
	// fmt.Printf("===> GCTL_ONCE_FILES=%#v\n", GlobalConfig.OnceFiles)
	// fmt.Printf("===> GCTL_PROTO_PATHS=%s\n", GlobalConfig.AllProtoPathList)
	// fmt.Printf("===> GCTL_PROTO_CENTRAL_REPO_PATH=%s\n", GlobalConfig.ProtoCentralRepoPath)
	// fmt.Printf("===> GCTL_DEFAULT_SERVICE_GROUP=%s\n", GlobalConfig.DefaultSvcGroup)
	// fmt.Printf("===> GCTL_DB_URI=%s\n", GlobalConfig.DbURI)
	// fmt.Printf("===> GCTL_ENABLE_ALLOC_PORT=%t\n", GlobalConfig.EnableAssignPort)
	// fmt.Printf("===> GCTL_ENABLE_ALLOC_ERRCODE=%t\n", GlobalConfig.EnableAssignErrcode)
	// fmt.Printf("===> GCTL_SVC_PORT_INTERVAL=%d\n", GlobalConfig.SvcPortInterval)
	// fmt.Printf("===> GCTL_SVC_ERRCODE_INTERVAL=%d\n", GlobalConfig.SvcErrcodeInterval)
	// fmt.Printf("===> GCTL_SVC_GROUP_INIT_PORT_MAP=%#v\n", GlobalConfig.SvcGroupInitPortMap)
	// fmt.Printf("===> GCTL_SVC_GROUP_INIT_ERRCODE_MAP=%#v\n", GlobalConfig.SvcGroupInitErrcodeMap)
}

func GetHomeDir() string {
	switch runtime.GOOS {
	case "windows":
		return os.Getenv("USERPROFILE")
	default:
		return os.Getenv("HOME")
	}
}
