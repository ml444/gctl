package cmd

import (
	"bytes"
	"fmt"
	"go/build"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"

	"github.com/ml444/gctl/parser"
	log "github.com/ml444/glog"
	"github.com/spf13/cobra"
)

var clientCmd = &cobra.Command{
	Use:     "client",
	Short:   "Generate client lib",
	Aliases: []string{"c"},
	Run: func(cmd *cobra.Command, args []string) {

		if protoPath == "" && len(args) == 0 {
			log.Error("You must provide the file of proto: gctl client -p=<protoFilepath> or gctl client <NAME>")
			return
		}
		if serviceGroup == "" && config.DefaultSvcGroup != "" {
			serviceGroup = config.DefaultSvcGroup
		}
		baseDir := config.TargetRootPath
		if protoPath == "" {
			protoPath = args[0]
			//protoPath = filepath.Join(baseDir, config.GoModulePrefix, fmt.Sprintf("%s.proto", arg))
		}
		protoPath = GetProtoAbsPath(protoPath)
		tmpDir := GetTemplateClientDir()
		onceFiles := config.OnceFiles
		log.Info("root location of code generation:", baseDir)
		log.Info("template path of code generation:", tmpDir)
		log.Info("files that are executed only once during initialization:", onceFiles)
		onceFileMap := map[string]bool{}
		for _, fileName := range onceFiles {
			onceFileMap[fileName] = true
		}
		var err error
		pd, err := parser.ParseProtoFile(protoPath)
		if err != nil {
			log.Errorf("err: %v", err)
			return
		}
		//dataMap["GoVersion"] = strings.TrimPrefix(runtime.Version(), "go")

		if config.EnableAssignErrcode {
			var moduleId int
			svcAssign := util.NewSvcAssign(
				config.DbDSN, getProtoName(protoPath), serviceGroup,
				config.SvcPortInterval, config.SvcErrcodeInterval,
				config.SvcGroupInitPortMap, config.SvcGroupInitErrcodeMap,
			)
			moduleId, err = svcAssign.GetModuleId()
			if err != nil {
				log.Error(err)
				return
			}
			pd.ModuleId = moduleId
		}

		clientRootDir := filepath.Join(baseDir, strings.Split(pd.Options["go_package"], ";")[0])
		err = filepath.Walk(tmpDir, func(path string, info fs.FileInfo, err error) error {
			if err != nil {
				log.Errorf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
				return err
			}
			if info.IsDir() {
				log.Warnf("skipping dir: %+v \n", info.Name())
				return nil
			}
			fileName := strings.TrimSuffix(info.Name(), config.TmplConfigFile.Template.FilesFormatSuffix)

			parentPath := strings.TrimRight(strings.TrimPrefix(path, tmpDir), info.Name())
			targetFile := clientRootDir + parentPath + fileName
			if util.IsFileExist(targetFile) && onceFileMap[fileName] {
				log.Infof("[%s] file is exist in this directory, skip it", targetFile)
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

		// generate protobuf file
		{
			if ok := checkProtoc(); !ok {
				return
			}
			log.Info("generating protobuf file")
			err = GenerateProtobuf(pd, baseDir, needGenGrpcPb)
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
		{
			pbFilepath := filepath.Join(absPath, fmt.Sprintf("%s.pb.go", pd.PackageName))
			areas, err := parser.ParsePbFile(pbFilepath, nil, nil)
			if err != nil {
				log.Fatal(err)
			}
			if err = parser.WritePbFile(pbFilepath, areas, false); err != nil {
				log.Fatal(err)
			}
		}

		// go mod tidy && go fmt
		{
			cmd := exec.Command("bash", "-c", "cd "+absPath+" && go mod tidy && go fmt ./...")
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

		time.Sleep(time.Millisecond * 100)
	},
}

func getProtoName(protoPath string) string {
	_, fname := filepath.Split(protoPath)
	return strings.TrimSuffix(fname, ".proto")
}

func checkProtoc() bool {
	p := exec.Command("protoc")
	if p.Run() != nil {
		log.Error("Please install protoc first and than rerun the command")
		switch runtime.GOOS {
		case "windows":
			log.Warn(
				`Install proto3.
https://github.com/google/protobuf/releases
Update protoc Go bindings via
> go get -u github.com/golang/protobuf/proto
> go get -u github.com/golang/protobuf/protoc-gen-go

See also
https://github.com/grpc/grpc-go/tree/master/examples`,
			)
		case "darwin":
			log.Warn(
				`Install proto3 from source macOS only.
> brew install autoconf automake libtool
> git clone https://github.com/google/protobuf
> ./autogen.sh ; ./configure ; make ; make install

Update protoc Go bindings via
> go get -u github.com/golang/protobuf/{proto,protoc-gen-go}

See also
https://github.com/grpc/grpc-go/tree/master/examples`,
			)
		default:
			log.Warn(`Install proto3
sudo apt-get install -y git autoconf automake libtool curl make g++ unzip
git clone https://github.com/google/protobuf.git
cd protobuf/
./autogen.sh
./configure
make
make check
sudo make install
sudo ldconfig # refresh shared library cache.`)
		}
		return false
	}
	return true
}
func GenerateProtobuf(pd *parser.ParseData, basePath string, needGenGrpcPb bool) error {
	var err error
	var args []string
	var protocName string
	var protoGenGoName string
	switch runtime.GOOS {
	case "windows":
		protocName = "protoc.exe"
		protoGenGoName = "protoc-gen-go.exe"
	default:
		protocName = "protoc"
		protoGenGoName = "protoc-gen-go"
	}
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		goPath = build.Default.GOPATH
	}
	protoGenGoPath := filepath.ToSlash(filepath.Join(goPath, "bin", protoGenGoName))
	args = append(args, fmt.Sprintf("--plugin=protoc-gen-go=%s", protoGenGoPath))
	args = append(args, fmt.Sprintf("--go_out=%s", filepath.ToSlash(basePath)))
	args = append(args, fmt.Sprintf("--validate_out=lang=go:%s", filepath.ToSlash(basePath)))
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
	return config.ThirdPartyProtoPath
}

func GetTemplateClientDir() string {
	var elems []string
	elems = append(elems, config.TmplRootDir)
	elems = append(elems, config.TmplConfigFile.Template.RelativeDir.Client...)
	return filepath.Join(elems...)
}
