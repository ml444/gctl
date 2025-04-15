package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ml444/gutil/osx"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/internal/db"
	"github.com/ml444/gctl/util"

	log "github.com/ml444/glog"
	"github.com/spf13/cobra"

	"github.com/ml444/gctl/parser"
)

func init() {
	clientCmd.Flags().StringVarP(&name, "proto", "p", "", "the filepath of proto")
	clientCmd.Flags().BoolVarP(&needGenGrpcPb, "grpc", "", true, "generate grpc pb file")
	clientCmd.Flags().StringVarP(&projectGroup, "service-group", "g", "", "a group of service, example: base|sys|biz...")
}

var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "Generate client lib",
	Aliases: []string{"c"},
	Run: func(_ *cobra.Command, args []string) {
		var err error
		err = RequiredParams(&name, args, &projectGroup)
		if err != nil {
			log.Error(err)
			return
		}
		tmplCfg := config.GlobalConfig.TemplatesConf
		protoFilePath := tmplCfg.ProtoTargetAbsPath(projectGroup, name)
		tmpDir := tmplCfg.ClientTmplAbsDir()
		log.Debug("template path of code generation: ", tmpDir)
		onceFiles := config.GlobalConfig.OnceFiles
		onceFileMap := map[string]bool{}
		for _, fileName := range onceFiles {
			onceFileMap[fileName] = true
		}
		pdCtx, err := parser.ParseProtoFile(protoFilePath)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		pdCtx.Group = projectGroup
		serviceName := getProtoName(name)
		pdCtx.Name = serviceName
		// pdCtx.GoPackage = tmplCfg.JoinGoPackage(projectGroup, serviceName)
		if config.GlobalConfig.EnableAssignErrcode {
			var moduleID int
			svcAssign, err := db.GetDispatcher(serviceName, projectGroup, &config.GlobalConfig)
			if err != nil {
				log.Error(err)
				return
			}
			defer svcAssign.Close()
			moduleID, err = svcAssign.GetModuleID()
			if err != nil {
				log.Error(err)
				return
			}
			pdCtx.ModuleID = moduleID
		}
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
				pdCtx.Ports = ports
			}
		}
		var clientRootDir string
		if pkgPath := pdCtx.Options["go_package"]; pkgPath != "" {
			if strings.Contains(pkgPath, ";") {
				pkgPath = strings.Split(pkgPath, ";")[0]
			}
			pdCtx.GoPackage = pkgPath
			clientRootDir = tmplCfg.ClientTargetAbsDir0(pkgPath)
		} else {
			clientRootDir = tmplCfg.ClientTargetAbsDir(projectGroup, serviceName)
		}
		pdCtx.ClientDir = clientRootDir
		log.Info("clientRootDir: ", clientRootDir)
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				log.Warnf("skipping dir: %+v \n", info.Name())
				return nil
			}
			targetFile := tmplCfg.ProcessFilePath(clientRootDir, path, tmpDir, info.Name())
			if util.IsFileExist(targetFile) && onceFileMap[tmplCfg.GetFileName(info.Name())] {
				log.Warnf("[%s] file is exist in this directory, skip it", targetFile)
				return nil
			}

			log.Infof("generating file: %s \n", targetFile)
			err = parser.GenerateTemplate(targetFile, path, pdCtx)
			if err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error walking the path %q: %v\n", tmpDir, err)
			return
		}

		// generate protobuf file
		{
			if ok := checkProtoc(); !ok {
				return
			}
			baseDir := config.GlobalConfig.TargetBaseDir

			log.Info("baseDir: ", baseDir)
			log.Info("GoPackage: ", pdCtx.GoPackage)
			log.Info("TargetBaseDir: ", config.GlobalConfig.TargetBaseDir)
			log.Info("root location of code generation: ", pdCtx.ClientDir)
			log.Info("generating protobuf file")
			err = GeneratePbFiles(pdCtx, baseDir, needGenGrpcPb)
			if err != nil {
				log.Error(err)
				return
			}
		}

		absPath, err := filepath.Abs(clientRootDir)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}

		// inject tag
		//{
		//	pbFilepath := filepath.Join(absPath, fmt.Sprintf("%s.pb.go", pd.PackageName))
		//	areas, err := parser.ParsePbFile(pbFilepath, nil, nil)
		//	if err != nil {
		//		log.Fatal(err)
		//	}
		//	if err = parser.WritePbFile(pbFilepath, areas, false); err != nil {
		//		log.Fatal(err)
		//	}
		//}

		// go mod tidy && go fmt
		if osx.IsFileExist(filepath.Join(absPath, "go.mod")) {
			util.CmdExec("cd " + absPath + " && go mod tidy")
			util.CmdExec("cd " + absPath + " && go fmt ./...")
		}
		util.CmdExec("cd " + absPath + " && goimports -w ./*.pb.go")
	},
}

func checkProtoc() bool {
	p := exec.Command("protoc")
	if p.Run() != nil {
		log.Error("Please install protoc first and than rerun the command")
		log.Warn("See (https://github.com/ml444/gctl/README.md#install-protoc)")
		return false
	}
	return true
}

func GeneratePbFiles(pdCtx *parser.CtxData, baseDir string, needGenGrpcPb bool) error {
	var err error
	var args []string
	var protocName string
	// var protoGenGoName string
	switch runtime.GOOS {
	case "windows":
		protocName = "protoc.exe"
	default:
		protocName = "protoc"
	}

	clientRelativePath, err := filepath.Rel(baseDir, pdCtx.ClientDir)
	if err != nil {
		log.Error(err)
		return err
	}
	dir, projectName := filepath.Split(config.GlobalConfig.TargetBaseDir)
	// baseDir := clientRootDir
	if pdCtx.GoPackage == "" {
		baseDir = pdCtx.ClientDir
	} else {
		if strings.HasPrefix(pdCtx.GoPackage, projectName) {
			baseDir = dir
		}
	}
	protoDir, protoName := filepath.Split(pdCtx.FilePath)
	log.Info("protoDir:", protoDir)
	log.Info("clientRelPath:", clientRelativePath)
	if len(config.GlobalConfig.ProtoPlugins) > 0 {
		args = append(args, config.GlobalConfig.ProtoPlugins...)
	} else {
		args = append(args, fmt.Sprintf("--go_out=%s", filepath.ToSlash(baseDir)))
		args = append(args, fmt.Sprintf("--go-grpc_out=%s", filepath.ToSlash(baseDir)))
		args = append(args, fmt.Sprintf("--go-http_out=%s", filepath.ToSlash(baseDir)))
		args = append(args, fmt.Sprintf("--go-gorm_out=%s", filepath.ToSlash(baseDir)))
		args = append(args, fmt.Sprintf("--go-errcode_out=%s", filepath.ToSlash(baseDir)))
		args = append(args, fmt.Sprintf("--go-validate_out=%s", filepath.ToSlash(baseDir)))
		args = append(args, fmt.Sprintf("--go-field_out=%s", filepath.ToSlash(baseDir)))
		// args = append(args, fmt.Sprintf("--go-field_out=include_prefix=Model:%s", filepath.ToSlash(baseDir)))
		if config.GlobalConfig.SwaggerCentralRepoPath != "" {
			swagDir := filepath.Join(config.GlobalConfig.SwaggerCentralRepoPath, pdCtx.Group)
			if !util.IsDirExist(swagDir) {
				os.MkdirAll(swagDir, 0o755)
			}

			args = append(args, fmt.Sprintf("--openapiv2_out=%s", swagDir))
		} else {
			args = append(args, fmt.Sprintf("--openapiv2_out=%s", filepath.ToSlash(pdCtx.ClientDir)))
		}
		// args = append(args, "--go_out=paths=source_relative:.")
		// args = append(args, "--go-grpc_out=paths=source_relative:.")
		// args = append(args, "--go-http_out=paths=source_relative:.")
		// args = append(args, "--go-gorm_out=paths=source_relative:.")
		// args = append(args, "--go-errcode_out=paths=source_relative:.")
		// args = append(args, "--go-validate_out=paths=source_relative:.")
		// args = append(args, "--go-field_out=paths=source_relative,include_prefix=Model:.")
		// args = append(args, "--openapiv2_out=.")
		args = append(args, fmt.Sprintf("--descriptor_set_out=%s_pb.descriptor", filepath.Join(clientRelativePath, pdCtx.Name)))
		args = append(args, "--include_source_info")
	}

	// include proto
	includePaths := getIncludePathList()
	for _, x := range includePaths {
		args = append(args, fmt.Sprintf("--proto_path=%s", x))
	}

	args = append(args, fmt.Sprintf("-I=%s", protoDir), protoName)
	// protocPath := filepath.ToSlash(filepath.Join(goPath, "bin", protocName))
	// cmd := exec.Command(protocPath, args...)
	cmd := exec.Command(protocName, args...)
	log.Info("exec:", cmd.String())

	var outBuf, errBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err = cmd.Run()
	outStr := outBuf.String()
	errStr := errBuf.String()
	if err != nil {
		log.Infof("Err: %s \nStdout: %s \nStderr: %s", err, outStr, errStr)
		return err
	}
	if outStr != "" {
		log.Info("out:", outStr)
	}
	if errStr != "" {
		log.Error("err:", errStr)
	}
	return nil
}

func getIncludePathList() []string {
	return config.GlobalConfig.ProtoPaths
}
