package cmd

import (
	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/internal/db"
	"github.com/ml444/gctl/util"

	"github.com/ml444/gkit/log"
	"github.com/spf13/cobra"

	"github.com/ml444/gctl/parser"
)

func init() {
	protoCmd.Flags().StringVarP(&name, "name", "n", "", "The name of proto")
	protoCmd.Flags().StringVarP(&projectGroup, "group", "g", "", "a group of service, example: base|sys|biz...")
}

var protoCmd = &cobra.Command{
	Use:     "proto",
	Short:   "init proto file",
	Aliases: []string{"p"},
	Run: func(_ *cobra.Command, args []string) {
		var err error
		err = RequiredParams(&name, args, &projectGroup)
		if err != nil {
			log.Error(err)
			return
		}
		tmplCfg, err := config.GetTmplFilesConf()
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}

		targetFilepath := tmplCfg.ProtoTargetAbsPath(projectGroup, name)
		if util.IsFileExist(targetFilepath) {
			log.Errorf("%s is existed", targetFilepath)
			return
		}
		protoName := getProtoName(name)
		pd := parser.CtxData{
			PackageName: protoName,
			GoPackage:   tmplCfg.JoinGoPackage(projectGroup, name),
			//ModulePrefix: config.JoinModulePrefixWithGroup(projectGroup),
		}
		pd.Group = projectGroup

		var firstErrcode = 1
		var endErrCode = 1 << 31
		if config.GlobalConfig.EnableAssignErrcode {
			var errCode int
			svcAssign, err := db.GetDispatcher(protoName, projectGroup, &config.GlobalConfig)
			if err != nil {
				log.Error(err)
				return
			}
			defer svcAssign.Close()

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
			tmplCfg.ProtoTmplAbsPath(),
			pd,
		)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Info("generate proto file success: ", targetFilepath)
	},
}
