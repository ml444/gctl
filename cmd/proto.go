package cmd

import (
	"fmt"
	"github.com/ml444/gctl/config"
	"github.com/ml444/gctl/util"
	"path/filepath"
	"strings"

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
		tmpPath := GetTemplateProtoDir()
		tmpName := config.TmplConfigFile.Template.ProtoFilename
		modulePrefix := JoinModulePrefixWithGroup()
		targetFilepath := GetProtoAbsPath(protoName)
		if util.IsFileExist(targetFilepath) {
			log.Errorf("%s is existed", targetFilepath)
			return
		}
		data := map[string]interface{}{
			"modulePrefix":    modulePrefix,
			"PackageName":     protoName,
			"ServiceName":     protoName,
			"CaseServiceName": fmt.Sprintf("%s%s", strings.ToTitle(protoName[:1]), protoName[1:]),
		}
		var firstErrcode int = 1
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
				firstErrcode = errCode + 1
			}
		}
		data["StartErrCode"] = firstErrcode

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
	},
}

const (
	ProtoFileSuffix = ".proto"
	ServiceNameVar  = "{SERVICE_NAME}"
)

func GetProtoAbsPath(protoName string) string {
	if filepath.IsAbs(protoName) {
		return protoName
	}
	var elems []string
	repoPath := config.ProtoCentralRepoPath
	if repoPath != "" {
		elems = append(elems, repoPath)
	} else {
		elems = append(elems, filepath.Join(config.TargetRootPath, config.GoModulePrefix))
	}
	if serviceGroup != "" {
		elems = append(elems, serviceGroup)
	}
	serviceName := protoName
	if strings.HasSuffix(protoName, ProtoFileSuffix) {
		serviceName = strings.TrimSuffix(serviceName, ProtoFileSuffix)
	} else {
		protoName = fmt.Sprintf("%s.proto", protoName)
	}
	for _, el := range config.TmplConfigFile.Target.RelativeDir.Proto {
		elems = append(elems, strings.ReplaceAll(el, ServiceNameVar, serviceName))
	}
	elems = append(elems, protoName)
	return filepath.Join(elems...)
}

func JoinModulePrefixWithGroup() string {
	modulePrefix := config.GoModulePrefix
	if serviceGroup != "" {
		return filepath.Join(modulePrefix, serviceGroup)
	}
	return modulePrefix
}

func GetTemplateProtoDir() string {
	var elems []string
	elems = append(elems, config.TmplRootDir)
	elems = append(elems, config.TmplConfigFile.Template.RelativeDir.Proto...)
	//elems = append(elems, templateConfigFile.TemplateProtoFilename)
	return filepath.Join(elems...)
}
