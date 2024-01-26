package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ml444/gkit/log"
	"github.com/spf13/cobra"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/parser"
	"github.com/ml444/gctl/util"
)

func init() {
	projectCmd.Flags().StringVarP(&projectName, "name", "n", "", "The name of proto")
	projectCmd.Flags().StringVarP(&projectGroup, "group", "g", "", "a group of service, example: base|sys|biz...")
}

var projectName string

var projectCmd = &cobra.Command{
	Use:     "project",
	Short:   "Generate project files by template",
	Aliases: []string{"p"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		tmplCfg, err := config.GetTmplFilesConf()
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}

		serviceName := getProtoName(protoPath)
		protoPath = tmplCfg.ProtoTargetAbsPath(projectGroup, protoPath)
		//baseDir := config.GlobalConfig.TargetBaseDir
		onceFiles := config.GlobalConfig.OnceFiles
		onceFileMap := map[string]bool{}
		for _, fileName := range onceFiles {
			onceFileMap[fileName] = true
		}
		pd, err := parser.ParseProtoFile(protoPath)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		pd.ModulePrefix = config.JoinModulePrefixWithGroup(projectGroup)

		clientTempDir := tmplCfg.ClientTmplAbsDir()
		protoTempPath := tmplCfg.ProtoTmplAbsPath()
		projectTempDir := tmplCfg.ProjectTmplAbsDir()
		projectRootDir := tmplCfg.ProjectTargetAbsDir(projectGroup, serviceName)
		log.Debug("project dir:", projectRootDir)
		log.Debug("template project dir:", projectTempDir)
		err = filepath.Walk(projectTempDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				log.Debugf("skipping a dir without errors: %+v \n", info.Name())
				return nil
			}
			if path == protoTempPath {
				log.Debugf("skipping proto file: %+v \n", path)
				return nil
			}
			if dir, _ := filepath.Split(path); strings.TrimSuffix(dir, string(os.PathSeparator)) == clientTempDir {
				log.Debugf("skipping client file: %+v \n", path)
				return nil
			} else {
				log.Infof("generating dir: %+v \n", dir)
			}

			fileName := strings.TrimSuffix(info.Name(), tmplCfg.TempFileExtSuffix())
			parentPath := strings.TrimSuffix(strings.TrimPrefix(path, projectTempDir), info.Name())
			targetFile := projectRootDir + parentPath + fileName
			if util.IsFileExist(targetFile) && onceFileMap[fileName] {
				log.Warnf("[%s] file is exist in this directory, skip it", targetFile)
				return nil
			}

			log.Infof("generating file: %s", targetFile)
			err = parser.GenerateTemplate(targetFile, path, pd)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", projectTempDir, err)
			return
		}

		// go mod tidy && go fmt
		{
			util.CmdExec("cd " + projectRootDir + " && go mod tidy && go fmt ./...")
		}
	},
}
