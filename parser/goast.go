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
func (pd *CtxData) ParseGoFile(filePath string) error {
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
func (pd *CtxData) parseServiceMethod(astFile *ast.File) {
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

type GoFileAST struct {
	Consts  map[string]struct{}
	Objects map[string]string
	Funcs   map[string]struct{}
	Structs map[string]*StructAST
}

type StructAST struct {
	StructName string
	Fields     map[string]string
	Methods    map[string]*MethodAST
}

type MethodAST struct {
	Name    string
	ReqArgs []string
	RspArgs []string
}

func ParseFile(filePath string) (*GoFileAST, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	tf := token.NewFileSet()
	astFile, err := parser.ParseFile(tf, filePath, src, parser.ParseComments)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}
	data := GoFileAST{
		Consts:  map[string]struct{}{},
		Objects: map[string]string{},
		Funcs:   map[string]struct{}{},
		Structs: map[string]*StructAST{},
	}

	{
		for k, obj := range astFile.Scope.Objects {
			data.Objects[k] = obj.Kind.String()
		}
	}
	for _, decl := range astFile.Decls {
		switch decl.(type) {
		case *ast.GenDecl:
			// TODO
		case *ast.FuncDecl:
			v := decl.(*ast.FuncDecl)
			if v.Recv != nil {
				// NOTE: this is method function

				// svcName := v.Recv.List[0].Type.(*ast.Ident).Name
				// svcName := v.Recv.List[0].Type.(*ast.StarExpr).X.(*ast.Ident).Name
				ident, ok := v.Recv.List[0].Type.(*ast.Ident)
				if !ok {
					continue
				}
				svcName := ident.Name
				methodName := v.Name.Name
				if methodName[0] >= 'A' && methodName[0] <= 'Z' {
					st, ok := data.Structs[svcName]
					if !ok {
						st = &StructAST{
							StructName: svcName,
						}
						data.Structs[svcName] = st
					}
					method := &MethodAST{
						Name: methodName,
					}
					for _, arg := range v.Type.Params.List {
						reqExpr := arg.Type.(*ast.StarExpr).X.(*ast.SelectorExpr)
						req := strings.Join([]string{reqExpr.X.(*ast.Ident).Name, reqExpr.Sel.Name}, ".")
						method.ReqArgs = append(method.ReqArgs, req)
					}
					for _, arg := range v.Type.Results.List {
						rspExpr := arg.Type.(*ast.StarExpr).X.(*ast.SelectorExpr)
						rsp := strings.Join([]string{rspExpr.X.(*ast.Ident).Name, rspExpr.Sel.Name}, ".")
						method.RspArgs = append(method.RspArgs, rsp)
					}
					st.Methods[methodName] = method
				}
			} else {
				// normal function
				data.Funcs[v.Name.Name] = struct{}{}
			}
		}
	}
	return &data, nil
}
