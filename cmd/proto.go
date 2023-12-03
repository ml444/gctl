package cmd

import (
	"strings"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"

	"github.com/ml444/gkit/log"
	"github.com/spf13/cobra"

	"github.com/ml444/gctl/parser"
)

var protoCmd = &cobra.Command{
	Use:     "proto",
	Short:   "init proto file",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if protoName == "" && len(args) == 0 {
			log.Error("proto name must be provided")
			return
		}
		err = config.InitTmplFilesConf()
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		if !validate(protoName) {
			return
		}
		if serviceGroup == "" {
			serviceGroup = config.GlobalConfig.DefaultSvcGroup
		}
		//modulePrefix := config.JoinModulePrefixWithGroup(serviceGroup)
		targetFilepath := config.GetTargetProtoAbsPath(serviceGroup, protoName)
		if util.IsFileExist(targetFilepath) {
			log.Errorf("%s is existed", targetFilepath)
			return
		}
		pd := parser.ParseData{
			PackageName:  protoName,
			ModulePrefix: config.JoinModulePrefixWithGroup(serviceGroup),
		}

		var firstErrcode = 1
		var endErrCode = 1 << 31
		if config.GlobalConfig.EnableAssignErrcode {
			var err error
			var errCode int
			svcAssign := util.NewSvcAssign(protoName, serviceGroup)
			err = svcAssign.GetOrAssignPortAndErrcode(nil, &errCode)
			if err != nil {
				log.Error(err)
				return
			}
			if errCode != 0 {
				firstErrcode = errCode
				endErrCode = errCode + config.GlobalConfig.SvcErrcodeInterval - 1
			}
		}
		pd.StartErrCode = firstErrcode
		pd.EndErrCode = endErrCode

		err = parser.GenerateTemplate(
			targetFilepath,
			config.GetTempProtoAbsPath(),
			pd,
		)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Info("generate proto file success: ", targetFilepath)
	},
}

func validate(name string) bool {
	if strings.Contains(name, "-") {
		log.Error("prohibited use of '-'")
		return false
	}
	return true
}
