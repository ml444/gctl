package cmd

import (
	"fmt"
	"io/fs"
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
	Run: func(_ *cobra.Command, args []string) {
		var err error
		err = RequiredParams(&projectName, args, &projectGroup)
		if err != nil {
			log.Error(err)
			return
		}

		ctx := parser.NewCtxData()
		ctx.Command = "project"
		ctx.Name = projectName
		ctx.Cfg = &config.GlobalConfig
		tmplCfg := ctx.Cfg.TemplatesConf
		projectTempDir := tmplCfg.ProjectTmplAbsDir()
		projectRootDir := tmplCfg.ProjectTargetAbsDir(projectGroup, projectName)
		log.Debug("project dir:", projectRootDir)
		log.Debug("template project dir:", projectTempDir)
		var genFileDescList [][2]string
		err = filepath.Walk(projectTempDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}

			if info.IsDir() {
				// if it is existed, don't generate the dir and files in it
				relativeDir := strings.TrimPrefix(path, projectTempDir)
				if relativeDir != "" && util.IsDirExist(projectRootDir+relativeDir) {
					return fmt.Errorf("%s is existed", projectRootDir+relativeDir)
				}
				return nil
			}

			fileName := strings.TrimSuffix(info.Name(), tmplCfg.TempFileExtSuffix())
			parentPath := strings.TrimSuffix(strings.TrimPrefix(path, projectTempDir), info.Name())
			targetFile := projectRootDir + parentPath + fileName
			if util.IsFileExist(targetFile) {
				log.Warnf("file is existed, skip it: %q \n", targetFile)
				return nil
			}
			genFileDescList = append(genFileDescList, [2]string{targetFile, path})
			return nil
		})
		if err != nil {
			log.Errorf("walking error: %v\n", err)
			return
		}
		for _, v := range genFileDescList {
			log.Infof("generating file: %s\n", v[0])
			err = parser.GenerateTemplate(v[0], v[1], ctx)
			if err != nil {
				log.Errorf("err: %v", err)
				return
			}
		}

		// go mod tidy && go fmt
		{
			//util.CmdExec("cd " + projectRootDir + " && go mod tidy && go fmt ./...")
		}
	},
}
