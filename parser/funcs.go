package parser

import (
	"path/filepath"
	"strings"
	"text/template"

	"github.com/ml444/gctl/util"
)

var funcMap = template.FuncMap{
	"Concat":       util.Concat,
	"TrimSpace":    strings.TrimSpace,
	"TrimPrefix":   strings.TrimPrefix,
	"HasPrefix":    strings.HasPrefix,
	"Contains":     strings.Contains,
	"ToUpper":      strings.ToUpper,
	"ToUpperFirst": util.ToUpperFirst,
	"ToLowerFirst": util.ToLowerFirst,
	"ToSnakeCase":  util.ToSnakeCase,
	"ToCamelCase":  util.ToCamelCase,
	"Add":          util.Add,
	// "GetStatusCodeFromComment": util.GetStatusCodeFromComment,
	"GoModule": GoModule,
}

func GoModule(ctx *CtxData, command string) string {
	goModulePrefix := goModulePrefix(ctx)
	switch command {
	case "project":
		if goModulePrefix != "" {
			return goModulePrefix + "/" + ctx.Name
		}
		return ctx.Name
	case "server":
		tmplCfg := ctx.Cfg.TemplatesConf
		// get project name
		projectName := projectName(ctx)
		var elems []string
		if goModulePrefix != "" {
			elems = append(elems, goModulePrefix)
		}
		elems = append(elems, projectName)
		elems = append(elems, tmplCfg.ServerRelativeDirs(ctx.Name)...)
		return strings.Join(elems, "/")
	case "client":
		tmplCfg := ctx.Cfg.TemplatesConf
		// get project name
		projectName := projectName(ctx)
		var elems []string
		if goModulePrefix != "" {
			elems = append(elems, goModulePrefix)
		}
		elems = append(elems, projectName)
		elems = append(elems, tmplCfg.ClientRelativeDirs(ctx.Name)...)
		return strings.Join(elems, "/")
	}
	if goModulePrefix != "" {
		return goModulePrefix + "/" + ctx.Name
	}
	return ctx.Name
}

func projectName(ctx *CtxData) string {
	tmplCfg := ctx.Cfg.TemplatesConf
	targetPath := tmplCfg.ServerTargetAbsDir(ctx.ProjectGroup, ctx.Name)
	relativeDir := tmplCfg.ServerRelativePath(ctx.Name)
	projectPath := strings.TrimRight(strings.TrimSuffix(targetPath, relativeDir), string(filepath.Separator))
	_, projectName := filepath.Split(projectPath)
	return projectName
}

func goModulePrefix(ctx *CtxData) string {
	modulePrefix := ctx.Cfg.GoModulePrefix
	group := ctx.Cfg.DefaultSvcGroup
	if ctx.ProjectGroup != "" {
		group = ctx.ProjectGroup
	}
	if modulePrefix != "" {
		return modulePrefix + "/" + group
	}
	return modulePrefix
}
