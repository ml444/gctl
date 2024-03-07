package cmd

import (
	"bytes"
	"fmt"
	"io/fs"
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
		name = tmplCfg.ProtoTargetAbsPath(projectGroup, name)
		tmpDir := tmplCfg.ClientTmplAbsDir()
		log.Debug("template path of code generation: ", tmpDir)
		onceFiles := config.GlobalConfig.OnceFiles
		onceFileMap := map[string]bool{}
		for _, fileName := range onceFiles {
			onceFileMap[fileName] = true
		}
		pd, err := parser.ParseProtoFile(name)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		serviceName := getProtoName(name)
		if config.GlobalConfig.EnableAssignErrcode {
			var moduleID int
			svcAssign, err := db.NewSvcAssign(serviceName, projectGroup, &config.GlobalConfig)
			if err != nil {
				log.Error(err)
				return
			}
			moduleID, err = svcAssign.GetModuleID()
			if err != nil {
				log.Error(err)
				return
			}
			pd.ModuleID = moduleID
		}
		var clientRootDir string
		if pkgPath := pd.Options["go_package"]; pkgPath != "" {
			if strings.Contains(pkgPath, ";") {
				pkgPath = strings.Split(pkgPath, ";")[0]
			}
			clientRootDir = tmplCfg.ClientTargetAbsDir0(pkgPath)
		} else {
			clientRootDir = tmplCfg.ClientTargetAbsDir(projectGroup, serviceName)
		}
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				log.Warnf("skipping dir: %+v \n", info.Name())
				return nil
			}
			fileName := strings.TrimSuffix(info.Name(), tmplCfg.TempFileExtSuffix())
			parentPath := strings.TrimRight(strings.TrimPrefix(path, tmpDir), info.Name())
			targetFile := clientRootDir + parentPath + fileName
			if util.IsFileExist(targetFile) && onceFileMap[fileName] {
				log.Warnf("[%s] file is exist in this directory, skip it", targetFile)
				return nil
			}

			log.Infof("generating file: %s \n", targetFile)
			err = parser.GenerateTemplate(targetFile, path, pd)
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
			if len(pd.Options["go_package"]) == 0 {
				baseDir = clientRootDir
			}
			log.Debug("root location of code generation: ", baseDir)
			log.Info("generating protobuf file")
			err = GeneratePbFiles(pd, baseDir, needGenGrpcPb)
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
func GeneratePbFiles(pd *parser.CtxData, basePath string, needGenGrpcPb bool) error {
	var err error
	var args []string
	var protocName string
	//var protoGenGoName string
	switch runtime.GOOS {
	case "windows":
		protocName = "protoc.exe"
		//protoGenGoName = "protoc-gen-go.exe"
	default:
		protocName = "protoc"
		//protoGenGoName = "protoc-gen-go"
	}
	// goPath := os.Getenv("GOPATH")
	// if goPath == "" {
	// 	goPath = build.Default.GOPATH
	// }
	//protoGenGoPath := filepath.ToSlash(filepath.Join(goPath, "bin", protoGenGoName))
	//args = append(args, fmt.Sprintf("--plugin=protoc-gen-go=%s", protoGenGoPath))
	args = append(args, fmt.Sprintf("--go_out=%s", filepath.ToSlash(basePath)))
	args = append(args, fmt.Sprintf("--go-http_out=%s", filepath.ToSlash(basePath)))
	args = append(args, fmt.Sprintf("--go-gorm_out=%s", filepath.ToSlash(basePath)))
	args = append(args, fmt.Sprintf("--go-errcode_out=%s", filepath.ToSlash(basePath)))
	args = append(args, fmt.Sprintf("--go-validate_out=%s", filepath.ToSlash(basePath)))
	args = append(args, fmt.Sprintf("--go-field_out=include_prefix=Model:%s", filepath.ToSlash(basePath)))
	//args = append(args, fmt.Sprintf("--openapi_out=paths=import:%s", filepath.ToSlash(basePath)))
	//args = append(args, "--openapi_out=paths=source_relative:.")
	if needGenGrpcPb {
		args = append(args, fmt.Sprintf("--go-grpc_out=%s", filepath.ToSlash(basePath)))
	}

	// include proto
	includePaths := getIncludePathList()
	for _, x := range includePaths {
		args = append(args, fmt.Sprintf("--proto_path=%s", x))
	}

	protoDir, protoName := filepath.Split(pd.FilePath)
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
