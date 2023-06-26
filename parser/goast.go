package parser

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	log "github.com/ml444/glog"
)

// ParseGoFile 从一个go文件从找出制定后缀的结构体，然后分析结构体里面的方法函数，并排除私有方法函数。
func (pd *ParseData) ParseGoFile(filePath string) error {
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	tf := token.NewFileSet()
	p, err := parser.ParseFile(tf, filePath, src, parser.ParseComments)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	// init,go file
	// {
	// 	var funcBodyMap = make(map[string]string)
	// 	for _, decl := range p.Decls {
	// 		switch v := decl.(type) {
	// 		case *ast.FuncDecl:
	// 			if v.Name.Name == "init" {
	// 				funcBodyMap["init"] = v.Body.String()
	// 			}
	// 		}
	// 	}
	// }
	_, name := filepath.Split(filePath)
	switch name {
	case "dao.go":
		// processing dao.go file
		{
			var objMap = make(map[string]string)
			for k, obj := range p.Scope.Objects {
				objMap[k] = obj.Kind.String()
			}
			pd.ObjectMap = objMap
		}
	case "service.go":
		// processing service.go file
		pd.parseServiceMethod(p)
	}

	return nil
}

// Parse service.go file: find the structure that specifies the suffix from a go file,
// analyze the method functions inside the structure, and exclude private method functions.
func (pd *ParseData) parseServiceMethod(astFile *ast.File) {
	var svcMap = make(map[string]map[string]struct {
		Req string
		Rsp string
	})
	for _, decl := range astFile.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			// TODO
		case *ast.FuncDecl:
			v := decl.(*ast.FuncDecl)
			if v.Recv != nil { // this is method function
				// svcName := v.Recv.List[0].Type.(*ast.Ident).Name
				// svcName := v.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
				ident, ok := v.Recv.List[0].Type.(*ast.Ident)
				if !ok {
					continue
				}
				svcName := ident.Name
				if strings.HasSuffix(svcName, "Service") {
					methodName := v.Name.Name
					if methodName[0] >= 'A' && methodName[0] <= 'Z' {
						if _, ok := svcMap[svcName]; !ok {
							svcMap[svcName] = make(map[string]struct {
								Req string
								Rsp string
							}, 0)
						}
						reqExpr := v.Type.Params.List[1].Type.(*ast.StarExpr).X.(*ast.SelectorExpr)
						rspExpr := v.Type.Results.List[0].Type.(*ast.StarExpr).X.(*ast.SelectorExpr)
						req := strings.Join([]string{reqExpr.X.(*ast.Ident).Name, reqExpr.Sel.Name}, ".")
						rsp := strings.Join([]string{rspExpr.X.(*ast.Ident).Name, rspExpr.Sel.Name}, ".")
						svcMap[svcName][methodName] = struct {
							Req string
							Rsp string
						}{Req: req, Rsp: rsp}
					}
				}
			}
		}
	}
	pd.ServiceMethodMap = svcMap
}
