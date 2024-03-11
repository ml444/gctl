package cmd

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/internal/db"

	"github.com/ml444/gctl/util"

	log "github.com/ml444/glog"
	"github.com/spf13/cobra"

	"github.com/ml444/gctl/parser"
)

func init() {
	serverCmd.Flags().StringVarP(&name, "proto", "p", "", "The file of proto")
	serverCmd.Flags().StringVarP(&projectGroup, "group", "g", "", "a group of service, example: base|sys|biz...")
}

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Generate server lib",
	Aliases: []string{"s"},
	Run: func(_ *cobra.Command, args []string) {
		var err error
		err = RequiredParams(&name, args, &projectGroup)
		if err != nil {
			log.Error(err)
			return
		}

		serviceName := getProtoName(name)
		tmplCfg := config.GlobalConfig.TemplatesConf
		protoPath := tmplCfg.ProtoTargetAbsPath(projectGroup, name)
		//baseDir := config.GlobalConfig.TargetBaseDir
		onceFiles := config.GlobalConfig.OnceFiles
		onceFileMap := map[string]bool{}
		for _, fileName := range onceFiles {
			onceFileMap[fileName] = true
		}
		ctx, err := parser.ParseProtoFile(protoPath)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		ctx.Name = serviceName
		ctx.Command = "server"
		ctx.Cfg = &config.GlobalConfig
		if config.GlobalConfig.EnableAssignPort {
			var port int
			svcAssign, err := db.GetDispatcher(serviceName, projectGroup, &config.GlobalConfig)
			if err != nil {
				log.Error(err)
				return
			}
			defer svcAssign.Close()
			err = svcAssign.GetOrAssignPortAndErrcode(&port, nil)
			if err != nil {
				log.Error(err)
				return
			}
			if port != 0 {
				var ports []int
				for i := 0; i < config.GlobalConfig.SvcPortInterval; i++ {
					ports = append(ports, port+i)
				}
				ctx.Ports = ports
			}
		}
		//clientTempDir := tmplCfg.ClientTmplAbsDir()
		//protoTempPath := tmplCfg.ProtoTmplAbsPath()
		serverTempDir := tmplCfg.ServerTmplAbsDir()
		// serverRootDir := filepath.Join(baseDir, fmt.Sprintf("%sServer", strings.Split(pd.Options["go_package"], ";")[0]))
		serverRootDir := tmplCfg.ServerTargetAbsPath(projectGroup, serviceName)
		log.Debug("server root dir:", serverRootDir)
		log.Debug("template root dir:", serverTempDir)
		err = filepath.Walk(serverTempDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				log.Debugf("skipping a dir without errors: %+v \n", info.Name())
				return nil
			}

			targetFile := tmplCfg.ProcessFilePath(serverRootDir, path, serverTempDir, info.Name())
			if util.IsFileExist(targetFile) && onceFileMap[tmplCfg.GetFileName(info.Name())] {
				log.Printf("[%s] file is exist in this directory, skip it", targetFile)
				return nil
			}

			log.Infof("generating file: %s", targetFile)
			err = parser.GenerateTemplate(targetFile, path, ctx)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", serverTempDir, err)
			return
		}

		// go mod tidy && go fmt
		{
			util.CmdExec("cd " + serverRootDir + " && go mod tidy && go fmt ./...")
		}
	},
}
