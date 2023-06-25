package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

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
		tmpPath := config.GetTemplateProtoDir()
		tmpName := config.TmplConfigFile.Template.ProtoFilename
		modulePrefix := config.JoinModulePrefixWithGroup(serviceGroup)
		targetFilepath := config.GetProtoAbsPath(serviceGroup, protoName)
		if util.IsFileExist(targetFilepath) {
			log.Errorf("%s is existed", targetFilepath)
			return
		}
		data := map[string]interface{}{
			"ModulePrefix":    modulePrefix,
			"PackageName":     protoName,
			"ServiceName":     protoName,
			"CaseServiceName": fmt.Sprintf("%s%s", strings.ToTitle(protoName[:1]), protoName[1:]),
		}
		var firstErrcode = 1
		var endErrCode = 1 << 31
		if config.EnableAssignErrcode {
			var err error
			var errCode int
			svcAssign := util.NewSvcAssign(
				config.DbDSN, protoName, serviceGroup,
				config.SvcPortInterval, config.SvcErrcodeInterval,
				config.SvcGroupInitPortMap, config.SvcGroupInitErrcodeMap,
			)
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
		data["StartErrCode"] = firstErrcode
		data["EndErrCode"] = endErrCode

		err := parser.GenerateTemplate(
			targetFilepath,
			filepath.Join(tmpPath, tmpName),
			tmpName,
			data,
			funcMap,
		)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Info("generate proto file success: ", targetFilepath)
	},
}
