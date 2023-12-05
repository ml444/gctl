package cmd

import (
	"fmt"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"

	"github.com/ml444/gkit/log"
	"github.com/spf13/cobra"

	"github.com/ml444/gctl/parser"
)

func init() {
	protoCmd.Flags().StringVarP(&protoPath, "name", "n", "", "The name of proto")
	protoCmd.Flags().StringVarP(&serviceGroup, "group", "g", "", "a group of service, example: base|sys|biz...")
}

var protoCmd = &cobra.Command{
	Use:     "proto",
	Short:   "init proto file",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		err = CheckAndInit(&protoPath, args, &serviceGroup)
		if err != nil {
			log.Error(err)
			return
		}
		fmt.Println(protoPath)
		targetFilepath := config.GetTargetProtoAbsPath(serviceGroup, protoPath)
		if util.IsFileExist(targetFilepath) {
			log.Errorf("%s is existed", targetFilepath)
			return
		}
		protoName := getProtoName(protoPath)
		pd := parser.ParseData{
			PackageName: protoName,
			GoPackage:   config.GetGoPackage(serviceGroup, protoPath),
			//ModulePrefix: config.JoinModulePrefixWithGroup(serviceGroup),
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
