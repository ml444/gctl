package cmd

import (
	"os"
	"strings"
	"text/template"
	"time"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"

	log "github.com/ml444/glog"
	"github.com/spf13/cobra"
)

var (
	debug bool

	needGenGrpcPb bool
	serviceGroup  string
	protoName     string
	protoPath     string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "If you want to se the debug logs.")
	protoCmd.Flags().StringVarP(&protoName, "name", "n", "", "The name of proto")
	protoCmd.Flags().StringVarP(&serviceGroup, "service-group", "g", "", "a group of service, example: base|sys|biz...")

	clientCmd.Flags().StringVarP(&protoPath, "proto", "p", "", "The filepath of proto")
	clientCmd.Flags().BoolVarP(&needGenGrpcPb, "grpc", "", true, "Select whether to generate xxx_grpc.pb file")
	clientCmd.Flags().StringVarP(&serviceGroup, "service-group", "g", "", "a group of service, example: base|sys|biz...")

	serverCmd.Flags().StringVarP(&protoPath, "proto", "p", "", "The filepath of proto")
	serverCmd.Flags().StringVarP(&serviceGroup, "service-group", "g", "", "a group of service, example: base|sys|biz...")
}

var rootCmd = &cobra.Command{
	Use:     "gctl",
	Example: "gctl <command> [command options]",
	Short:   "A code generation and checking tool for Go microservices",
	Run: func(cmd *cobra.Command, args []string) {
		err := cmd.Help()
		if err != nil {
			log.Errorf("err: %v", err)
		}
		if debug {
			config.PrintImportantVars()
		}
	},
}

// Execute runs the root command
func Execute() {
	var err error

	config.SetDefaults()
	err = config.InitGlobalVar()
	if err != nil {
		println(err.Error())
		return
	}

	rootCmd.AddCommand(clientCmd, serverCmd, protoCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	time.Sleep(time.Millisecond * 100)
}

var funcMap = template.FuncMap{
	"Concat":                   util.Concat,
	"TrimSpace":                strings.TrimSpace,
	"TrimPrefix":               strings.TrimPrefix,
	"HasPrefix":                strings.HasPrefix,
	"Contains":                 strings.Contains,
	"GetStatusCodeFromComment": util.GetStatusCodeFromComment,
	"ToUpper":                  strings.ToUpper,
	"ToUpperFirst":             util.ToUpperFirst,
	"ToLowerFirst":             util.ToLowerFirst,
	"CamelToSnake":             util.CamelToSnake,
	"SnakeToCamel":             util.SnakeToCamel,
	"Add":                      util.Add,
}

//func init() {
//	rootCmd.PersistentFlags().BoolP("debug", "d", false, "If you want to se the debug logs.")
//	rootCmd.PersistentFlags().BoolP("force", "f", false, "Force overide existing files without asking.")
//	rootCmd.PersistentFlags().StringP("folder", "b", "", "If you want to specify the base folder of the project.")
//	viper.BindPFlag("gctl_folder", rootCmd.PersistentFlags().Lookup("folder"))
//	viper.BindPFlag("gctl_force", rootCmd.PersistentFlags().Lookup("force"))
//	viper.BindPFlag("gctl_debug", rootCmd.PersistentFlags().Lookup("debug"))
//}
