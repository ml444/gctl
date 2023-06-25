package parser

import (
	"runtime"
	"strings"

	"github.com/emicklei/proto"
)

type ParseData struct {
	GoVersion    string
	Ports        []int
	StartErrCode int
	ModuleId     int

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
	CmdID    string
	Url      string
	Flags    string
	UserType string
	PermList []string
	Options  map[string]string

	CommentLines []string
	//commentMap   map[string]*linesCommentNode
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
