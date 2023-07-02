package parser

import (
	"errors"
	"runtime"
	"strings"

	"github.com/emicklei/proto"
)

type ParseData struct {
	GoVersion    string
	Ports        []int
	StartErrCode int
	EndErrCode   int
	ModuleId     int
	ModulePrefix string

	// from proto file
	FilePath         string
	PackageName      string
	Options          map[string]string
	ImportMap        map[string]bool
	ServiceImportMap map[string][]string
	ListOptionMap    map[string]*ListReqOption
	ServiceList      []*Service
	ErrCodeList      []Enum
	ModelList        []*proto.Message
	MessageList      []*Message
	ModelFieldList   []string

	// from go file
	FileName         string
	ServiceMethodMap map[string]map[string]struct {
		Req string
		Rsp string
	}
	ObjectMap   map[string]string
	FuncBodyMap map[string]string
}

type Service struct {
	ServiceName string
	RpcList     []*RpcMethod
}

type RpcMethod struct {
	Name         string
	RequestType  string
	ResponseType string
	//StreamRequest	bool
	//StreamResponse	bool
	//Comment
	Options map[string]string

	CommentLines []string
	//commentMap   map[string]*linesCommentNode
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

func NewParseData() *ParseData {
	v := strings.TrimPrefix(runtime.Version(), "go")
	vList := strings.Split(v, ".")
	if len(vList) >= 3 {
		v = strings.Join(vList[:2], ".")
	}
	return &ParseData{
		GoVersion: v,
	}
}

func (pd *ParseData) GetFirstGoPackage() (string, error) {
	list := strings.Split(pd.Options["go_package"], ";")
	if len(list) == 0 {
		return "", errors.New("no go_package in proto file")
	}
	return list[0], nil
}
