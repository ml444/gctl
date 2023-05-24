package config

import (
	"github.com/spf13/viper"
	"os"
)

const (
	KeyTemplateRootDir     = "TEMPLATES_ROOT_DIR"
	KeyTargetRootPath      = "TARGET_ROOT_PATH"
	KeyModulePrefix        = "MODULE_PREFIX"
	KeyOnceFiles           = "ONCE_FILES"
	KeyThirdPartyProtoPath = "THIRD_PARTY_PROTO_PATH"

	KeyProtoCentralRepoPath = "PROTO_CENTRAL_REPO_PATH" // Store proto files centrally in one repository
	KeyDefaultServiceGroup  = "DEFAULT_SERVICE_GROUP"

	KeyDbDSN                  = "DB_DSN"
	KeySvcGroupInitPortMap    = "SVC_GROUP_INIT_PORT_MAP"
	KeySvcPortInterval        = "SVC_PORT_INTERVAL" // Port interval between services
	KeyEnableAssignPort       = "ENABLE_ALLOC_PORT"
	KeySvcGroupInitErrcodeMap = "SVC_GROUP_INIT_ERRCODE_MAP"
	KeySvcErrcodeInterval     = "SVC_ERRCODE_INTERVAL" // Errcode range of the service
	KeyEnableAssignErrcode    = "ENABLE_ALLOC_ERRCODE"
)

func init() {
	viper.SetEnvPrefix("gctl")
	viper.AutomaticEnv()
}

func SetDefaults() {
	viper.SetDefault(KeyTargetRootPath, os.Getenv("HOME"))
	viper.SetDefault(KeyOnceFiles, ".gitignore go.mod .editorconfig README.md Dockerfile Makefile")
	viper.SetDefault(KeySvcPortInterval, 5)
	viper.SetDefault(KeyEnableAssignPort, false)
	viper.SetDefault(KeySvcErrcodeInterval, 1000)
	viper.SetDefault(KeyEnableAssignErrcode, false)
}
