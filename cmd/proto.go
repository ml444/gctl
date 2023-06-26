package cmd

import (
	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"

	"github.com/ml444/gctl/parser"
	log "github.com/ml444/glog"
	"github.com/spf13/cobra"
)

var protoCmd = &cobra.Command{
	Use:     "proto",
	Short:   "init proto file",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		if protoName == "" {
			log.Error("proto name must be input:[-n=xxx]")
			return
		}
		if serviceGroup == "" {
			serviceGroup = config.DefaultSvcGroup
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
		if config.EnableAssignErrcode {
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
				endErrCode = errCode + config.SvcErrcodeInterval - 1
			}
		}
		pd.StartErrCode = firstErrcode
		pd.EndErrCode = endErrCode

		err := parser.GenerateTemplate(
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
