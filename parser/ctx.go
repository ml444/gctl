package parser

import (
	"errors"
	"runtime"
	"strings"

	"github.com/emicklei/proto"

	"github.com/ml444/gctl/config"
)

type CtxData struct {
	Cfg          *config.Config
	Command      string
	GoVersion    string
	Ports        []int
	StartErrCode int
	EndErrCode   int
	ModuleID     int
	// ModulePrefix string
	// GoModule string
	Name      string
	Group     string
	ClientDir string

	// from proto file
	FilePath         string
	PackageName      string
	GoPackage        string
	Options          map[string]string
	ImportMap        map[string]bool
	ServiceImportMap map[string][]string
	ListOptionMap    map[string]*ListReqOption
	ServiceList      []*Service
	ErrCodeList      []Enum
	ModelList        []*proto.Message
	MessageList      []*Message
	ModelFieldList   []string

	// Deprecated: from go file
	ServiceMethodMap map[string]map[string]struct {
		Req string
		Rsp string
	}
	// Deprecated: from go file
	ObjectMap map[string]string
	// Deprecated: from go file
	FuncBodyMap map[string]string

	// parse the specified go file
	ExistedGoFile *GoFileAST
}

type Service struct {
	ServiceName string
	RpcList     []*RpcMethod
}

type RpcMethod struct {
	Name         string
	RequestType  string
	ResponseType string
	//  StreamRequest	bool
	//  StreamResponse	bool
	//  Comment
	Options map[string]string

	CommentLines []string
	//  commentMap   map[string]*linesCommentNode
}

type Enum struct {
	Name          string
	EnumFieldList []*proto.EnumField
}

type Message struct {
	Name      string
	FieldList []*MessageField
}

type MessageField struct {
	*proto.MapField
	*proto.NormalField
	*Message
}

type ListReqOption struct {
	ReqName      string // the name of request
	EnumName     string
	EnumFieldMap map[string]string
}

func NewCtxData() *CtxData {
	v := strings.TrimPrefix(runtime.Version(), "go")
	vList := strings.Split(v, ".")
	if len(vList) >= 3 {
		v = strings.Join(vList[:2], ".")
	}
	return &CtxData{
		GoVersion: v,
	}
}

func (pd *CtxData) GetFirstGoPackage() (string, error) {
	list := strings.Split(pd.Options["go_package"], ";")
	if len(list) == 0 {
		return "", errors.New("no go_package in proto file")
	}
	return list[0], nil
}
