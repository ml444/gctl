package parser

import (
	"github.com/ml444/gctl/util"
	log "github.com/ml444/glog"
	"github.com/spf13/viper"
	"path/filepath"
	"strings"
	"text/template"
)

func ParseTemplateToFile(pd *ProtoData, basePath, tempDir, tempName string, funcMap template.FuncMap) error {
	fPath := filepath.Join(basePath, pd.Options["go_package"], strings.TrimSuffix(tempName, viper.GetString("template_format_suffix")))
	return GenerateTemplate(fPath, tempDir, tempName, pd, funcMap)
}

func GenerateTemplate(fPath, tempFile, tempName string, data interface{}, funcMap template.FuncMap) error {
	f, err := util.OpenFile(fPath)
	if err != nil {
		return err
	}
	temp := template.New(tempName)
	if funcMap != nil {
		temp.Funcs(funcMap)
	}
	temp, err = temp.ParseFiles(tempFile)
	if err != nil {
		log.Error(err)
		return err
	}
	err = temp.Execute(f, data)
	if err != nil {
		log.Printf("Can't generate file %s,Error :%v\n", fPath, err)
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return nil
}
