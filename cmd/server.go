package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"

	"github.com/ml444/gctl/parser"
	log "github.com/ml444/glog"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:     "server",
	Short:   "Generate server lib",
	Aliases: []string{"s"},
	Run: func(cmd *cobra.Command, args []string) {
		var err error
		if protoPath == "" && len(args) == 0 {
			log.Error("You must provide the file of proto: gctl server -p=<protoFilepath> or gctl server <NAME>")
			return
		}
		if protoPath == "" {
			protoPath = args[0]
		}
		if serviceGroup == "" && config.DefaultSvcGroup != "" {
			serviceGroup = config.DefaultSvcGroup
		}
		protoPath = GetProtoAbsPath(protoPath)
		baseDir := config.TargetRootPath
		tmpDir := GetTemplateServerDir()
		onceFiles := config.OnceFiles
		log.Info("root location of code generation:", baseDir)
		log.Info("template path of code generation:", tmpDir)
		log.Info("files that are executed only once during initialization:", onceFiles)
		onceFileMap := map[string]bool{}
		for _, fileName := range onceFiles {
			onceFileMap[fileName] = true
		}
		pd, err := parser.ParseProtoFile(protoPath)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		if config.EnableAssignPort {
			var port int
			svcAssign := util.NewSvcAssign(
				config.DbDSN, pd.PackageName, serviceGroup,
				config.SvcPortInterval, config.SvcErrcodeInterval,
				config.SvcGroupInitPortMap, config.SvcGroupInitErrcodeMap,
			)
			err = svcAssign.GetOrAssignPortAndErrcode(&port, nil)
			if err != nil {
				log.Error(err)
				return
			}
			if port != 0 {
				var ports []int
				for i := 0; i < config.SvcPortInterval; i++ {
					ports = append(ports, port+i)
				}
				pd.Ports = ports
			}
		}
		tempFileSuffix := config.TmplConfigFile.Template.FilesFormatSuffix
		serverRootDir := filepath.Join(baseDir, fmt.Sprintf("%sServer", strings.Split(pd.Options["go_package"], ";")[0]))
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				return nil
			}
			fileName := strings.TrimSuffix(info.Name(), tempFileSuffix)
			parentPath := strings.TrimRight(strings.TrimPrefix(path, tmpDir), info.Name())
			targetFile := serverRootDir + parentPath + fileName
			if util.IsFileExist(targetFile) && onceFileMap[fileName] {
				log.Printf("[%s] file is exist in this directory, skip it", targetFile)
				return nil
			}

			log.Infof("generating file: %s", targetFile)
			err = parser.GenerateTemplate(targetFile, path, info.Name(), pd, funcMap)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", tmpDir, err)
			return
		}

		// go mod tidy && go fmt
		{
			cmd := exec.Command("bash", "-c", "cd "+serverRootDir+" && go mod tidy && go fmt ./...")
			log.Infof("exec: %s", cmd.String())
			var outBuf, errBuf bytes.Buffer
			cmd.Stdout = &outBuf
			cmd.Stderr = &errBuf
			err = cmd.Run()
			if err != nil {
				log.Infof("Err: %s ", err.Error())
				log.Info("Stdout: ", outBuf.String())
				log.Info("Stderr: ", errBuf.String())
				return
			}
			log.Infof(" fmt files: %s", outBuf.String())
		}
	},
}

func GetTemplateServerDir() string {
	var elems []string
	elems = append(elems, config.TmplRootDir)
	elems = append(elems, config.TmplConfigFile.Template.RelativeDir.Server...)
	return filepath.Join(elems...)
}
