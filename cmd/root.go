package cmd

import (
	"github.com/ml444/glog/level"
	"os"
	"time"

	"github.com/ml444/gctl/config"
	log "github.com/ml444/glog"
	logConf "github.com/ml444/glog/config"
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
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "If you want to se the debug logs, or print all environment variables.")
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

	if debug {
		err = log.InitLog(logConf.SetLevel2Logger(level.DebugLevel))
		if err != nil {
			println(err.Error())
		}
	}

	rootCmd.AddCommand(clientCmd, serverCmd, protoCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	time.Sleep(time.Millisecond * 100)
}

//func init() {
//	rootCmd.PersistentFlags().BoolP("debug", "d", false, "If you want to se the debug logs.")
//	rootCmd.PersistentFlags().BoolP("force", "f", false, "Force overide existing files without asking.")
//	rootCmd.PersistentFlags().StringP("folder", "b", "", "If you want to specify the base folder of the project.")
//	viper.BindPFlag("gctl_folder", rootCmd.PersistentFlags().Lookup("folder"))
//	viper.BindPFlag("gctl_force", rootCmd.PersistentFlags().Lookup("force"))
//	viper.BindPFlag("gctl_debug", rootCmd.PersistentFlags().Lookup("debug"))
//}
