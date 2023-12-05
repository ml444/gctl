package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ml444/glog/level"

	"github.com/ml444/gctl/config"

	log "github.com/ml444/glog"
	logConf "github.com/ml444/glog/config"
	"github.com/spf13/cobra"
)

var (
	debug bool

	needGenGrpcPb bool
	serviceGroup  string
	protoPath     string
)

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "If you want to se the debug logs, or print all environment variables.")

	//serverCmd.Flags().StringVarP(&protoPath, "proto", "p", "", "The filepath of proto")
	//serverCmd.Flags().StringVarP(&serviceGroup, "service-group", "g", "", "a group of service, example: base|sys|biz...")
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

func CheckAndInit(protoPath *string, args []string, serviceGroup *string) error {
	var err error
	if protoPath == nil || *protoPath == "" {
		if len(args) == 0 {
			return errors.New("proto name must be provided")
		}
		*protoPath = strings.TrimSpace(args[0])
	}

	if strings.Contains(*protoPath, "-") {
		return errors.New("prohibited use of '-'")
	}
	err = config.InitTmplFilesConf()
	if err != nil {
		log.Errorf("err: %v", err)
		return err
	}

	if serviceGroup == nil {
		*serviceGroup = config.GlobalConfig.DefaultSvcGroup
	}
	return nil
}
func getProtoName(protoPath string) string {
	_, fname := filepath.Split(protoPath)
	return strings.TrimSuffix(fname, ".proto")
}
