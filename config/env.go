package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	templateRootDir     = "TEMPLATES_ROOT_DIR"
	targetRootPath      = "TARGET_ROOT_PATH"
	modulePrefix        = "MODULE_PREFIX"
	onceFiles           = "ONCE_FILES"
	thirdPartyProtoPath = "THIRD_PARTY_PROTO_PATH"

	protoCentralRepoPath = "PROTO_CENTRAL_REPO_PATH" // Store proto files centrally in one repository
	defaultServiceGroup  = "DEFAULT_SERVICE_GROUP"

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
	viper.SetDefault(targetRootPath, os.Getenv("HOME"))
	viper.SetDefault(onceFiles, ".gitignore go.mod .editorconfig README.md")
	viper.SetDefault(thirdPartyProtoPath, filepath.Join(TmplRootDir, "protofiles")) // TODO: protoFile
	viper.SetDefault(KeySvcPortInterval, 5)
	viper.SetDefault(KeyEnableAssignPort, false)
	viper.SetDefault(KeySvcErrcodeInterval, 1000)
	viper.SetDefault(KeyEnableAssignErrcode, true)
}
