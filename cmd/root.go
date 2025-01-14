package cmd

import (
	"os"
	"time"

	"github.com/ml444/glog/level"

	"github.com/ml444/gctl/config"

	log "github.com/ml444/glog"
	"github.com/spf13/cobra"
)

var (
	debug bool

	needGenGrpcPb bool
	projectGroup  string
	name          string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "If you want to se the debug logs, or print all environment variables.")

	// serverCmd.Flags().StringVarP(&protoPath, "proto", "p", "", "The filepath of proto")
	// serverCmd.Flags().StringVarP(&projectGroup, "service-group", "g", "", "a group of service, example: base|sys|biz...")
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

	err = config.InitGlobalVar()
	if err != nil {
		println(err.Error())
		return
	}

	if debug {
		err = log.InitLog(log.SetLoggerLevel(level.DebugLevel))
		if err != nil {
			println(err.Error())
		}
	}

	rootCmd.AddCommand(projectCmd, clientCmd, serverCmd, protoCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
	time.Sleep(time.Millisecond * 100)
}
