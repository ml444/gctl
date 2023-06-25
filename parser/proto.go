package parser

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"

	"github.com/ml444/gctl/util"

	"github.com/emicklei/proto"
	log "github.com/ml444/glog"
)

// ParseProtoFile 解析proto文件
func ParseProtoFile(protoFilepath string) (*ParseData, error) {
	reader, err := os.Open(protoFilepath)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	defer func() {
		err := reader.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	protoData := NewParseData()
	protoData.FilePath = protoFilepath
	handlePackage := func(p *proto.Package) {
		protoData.PackageName = p.Name
	}

	handleOptions := func(o *proto.Option) {
		if protoData.Options == nil {
			protoData.Options = make(map[string]string)
		}
		protoData.Options[o.Name] = o.Constant.Source
	}

	handleService := func(s *proto.Service) {
		svc := &Service{
			ServiceName: s.Name,
			//RpcList:     nil,
		}

		protoData.ServiceList = append(protoData.ServiceList, svc)
		log.Info("===> serviceName:", s.Name)
	}

	handleRpcMethod := func(m *proto.RPC) {
		cmdID := 0
		url := ""
		flags := 0
		userType := ""
		perm := ""
		options := map[string]string{}
		log.Infof("=====> rpcName: %s", m.Name)
		//for _, opt := range m.Elements {
		//	v := &ProtoVisitor{}
		//	opt.Accept(v)
		//	if v.CmdID > 0 {
		//		cmdID = int(v.CmdID)
		//	}
		//	if v.Url != "" {
		//		url = v.Url
		//	}
		//	if v.Flags > 0 {
		//		flags = int(v.Flags)
		//	}
		//	if v.UserType != "" {
		//		userType = v.UserType
		//	}
		//	if v.Perm != "" {
		//		perm = v.Perm
		//	}
		//	if v.options != nil {
		//		for key, val := range v.options {
		//			options[key] = val
		//		}
		//	}
		//}

		var permList []string
		permItems := strings.Split(perm, "|")
		for _, v := range permItems {
			v = strings.TrimSpace(v)
			if v != "" {
				permList = append(permList, v)
			}
		}

		//if cmdID == 0 {
		//	log.Fatalf("method `%s` missed CmdID option", m.Name)
		//}

		method := &RpcMethod{
			Name:         m.Name,
			RequestType:  m.RequestType,
			ResponseType: m.ReturnsType,
			CmdID:        strconv.Itoa(cmdID),
			Url:          url,
			Flags:        strconv.Itoa(flags),
			PermList:     permList,
			UserType:     userType,
			Options:      options,
		}

		if m.Comment != nil {
			method.CommentLines = m.Comment.Lines
		}

		parent := &ProtoVisitor{}
		m.Parent.Accept(parent)
		for _, svc := range protoData.ServiceList {
			if parent.Name == svc.ServiceName {
				svc.RpcList = append(svc.RpcList, method)
			}
		}

	}

	handleEnum := func(e *proto.Enum) {
		if strings.HasSuffix(e.Name, "ErrCode") {
			var enum ProtoVisitor
			for _, ei := range e.Elements {
				ei.Accept(&enum)
			}
			protoData.ErrCodeList = append(
				protoData.ErrCodeList,
				Enum{Name: e.Name, EnumFieldList: enum.EnumFieldList})
		} else if strings.HasPrefix(e.Name, "ListOpt") {
			if protoData.ListOptionMap == nil {
				protoData.ListOptionMap = map[string]*ListReqOption{}
			}
			msg := Message{}
			vv := &ProtoVisitor{Message: &msg}
			e.Parent.Accept(vv)
			reqName := msg.Name
			if reqName != "" {
				for _, el := range e.Elements {
					el.Accept(vv)
				}

				listReqOpt := ListReqOption{
					ReqName:  reqName,
					EnumName: e.Name,
					//EnumFieldMap: map[string]string{},
				}

				fieldMap := map[string]string{}
				for _, v := range vv.EnumFieldList {
					if v.Comment == nil {
						continue
					}
					for _, line := range v.Comment.Lines {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "@valueType:") {
							lineV := strings.TrimSpace(strings.TrimPrefix(line, "@valueType:"))
							if vLen := len(lineV); vLen < 4 || vLen > 11 {
								log.Warnf("this valueType dose not match the requirements: %s", line)
								continue
							}
							fieldMap[v.Name] = lineV
							break
						}
					}
				}
				if len(fieldMap) > 0 {
					listReqOpt.EnumFieldMap = fieldMap
				}
				if protoData.ListOptionMap == nil {
					protoData.ListOptionMap = map[string]*ListReqOption{
						reqName: &listReqOpt,
					}
				} else {
					protoData.ListOptionMap[reqName] = &listReqOpt
				}
			}
		}
	}

	handleImport := func(i *proto.Import) {
		if protoData.ImportMap == nil {
			protoData.ImportMap = make(map[string]bool)
		}
		if protoData.ImportMap[i.Filename] {
			return
		}
		defer func() {
			protoData.ImportMap[i.Filename] = true
		}()

		if strings.HasPrefix(i.Filename, "google/") {
			return
		}

		// Determine whether the path exists
		pb := SearchImportPb(i.Filename)
		if pb != "" {
			//old := CurrentPb
			//
			//SetCurrentPb(i.Filename)
			//pdImport := ParsePbOrDie(pb)
			//
			//if pdImport.GoPackageName != "" {
			//	pd.ImportList = append(
			//		pd.ImportList, &ImportNode{
			//			ImportPath: i.Filename, GoPackage: pdImport.GoPackageName})
			//}
			//
			//SetCurrentPb(old)
		} else {
			log.Warnf("not found '%s' in current dir", i.Filename)
		}
	}

	handleMessage := func(m *proto.Message) {
		Msg := &Message{Name: m.Name}
		vv := &ProtoVisitor{Message: Msg}
		for _, v := range m.Elements {
			v.Accept(vv)
		}
		//pbList = append(pbList, pbMsg)
		//
		//key := fmt.Sprintf("%s_%s", CurrentMod, m.Name)
		//
		//if e, ok := PbMap[key]; ok {
		//	e.NameDupCnt++
		//} else {
		//	PbMap[key] = pbMsg
		//}

		if strings.HasPrefix(m.Name, "Model") {
			protoData.ModelList = append(protoData.ModelList, m)
		} else if strings.HasPrefix(m.Name, "List") && strings.HasSuffix(m.Name, "Req") {
			if len(vv.EnumFieldList) > 0 {
				log.Info(vv.EnumFieldList)
			}
		}
	}

	proto.Walk(definition,
		proto.WithImport(handleImport),
		proto.WithOption(handleOptions),
		proto.WithPackage(handlePackage),
		proto.WithService(handleService),
		proto.WithRPC(handleRpcMethod),
		proto.WithEnum(handleEnum),
		proto.WithMessage(handleMessage))

	// Make sure packageName is not empty
	if protoData.PackageName == "" {
		protoData.PackageName = strings.TrimSuffix(protoData.FilePath, ".proto")
	}
	return protoData, nil
}

func SearchImportPb(impPath string) string {
	if util.IsFileExist(impPath) {
		return impPath
	}

	//for _, incPath := range c.IncludePathList {
	//	p := strings.Join([]string{incPath, impPath}, utils.Sep)
	//	if IsFileExist(p) {
	//		return p
	//	}
	//}

	return ""
}

func Struct2map(data interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(data)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(b, &m)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return m, nil
}

// type ProtoData struct {
// 	GoVersion    string
// 	Ports        []int
// 	StartErrCode int
// 	ModuleId     int

// 	FilePath         string
// 	PackageName      string
// 	Options          map[string]string
// 	ImportMap        map[string]bool
// 	ServiceImportMap map[string][]string
// 	ListOptionMap    map[string]*ListReqOption
// 	ServiceList      []*Service
// 	ErrCodeList      []Enum
// 	ModelList        []*proto.Message
// 	MessageList      []*Message
// }

// type Service struct {
// 	ServiceName string
// 	RpcList     []*RpcMethod
// }

// type RpcMethod struct {
// 	Name         string
// 	RequestType  string
// 	ResponseType string
// 	//StreamRequest	bool
// 	//StreamResponse	bool
// 	//Comment
// 	CmdID    string
// 	Url      string
// 	Flags    string
// 	UserType string
// 	PermList []string
// 	Options  map[string]string

// 	CommentLines []string
// 	//commentMap   map[string]*linesCommentNode
// }

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

type ProtoVisitor struct {
	//proto.Visitor
	//proto.RPC
	Name          string
	EnumFieldList []*proto.EnumField
	Message       *Message
}

func (p *ProtoVisitor) VisitMessage(m *proto.Message) {
	if p.Message == nil {
		p.Message = &Message{Name: m.Name}
	} else {
		p.Message.Name = m.Name
	}
}

func (p *ProtoVisitor) VisitService(v *proto.Service) {
	p.Name = v.Name
}
func (p *ProtoVisitor) VisitSyntax(s *proto.Syntax) {
}
func (p *ProtoVisitor) VisitPackage(pkg *proto.Package) {
}

func (p *ProtoVisitor) VisitOption(o *proto.Option) {
	//if strings.Index(o.Name, "ext.CmdID") >= 0 {
	//	strVal := o.Constant.Source
	//	val, err := strconv.Atoi(strVal)
	//	if err != nil || val <= 0 {
	//		log.Fatalf("invalid CmdID `%s` at line %d", strVal, o.Position.Line)
	//	}
	//	p.CmdID = uint32(val)
	//} else if strings.Index(o.Name, "ext.Flags") >= 0 {
	//	strVal := o.Constant.Source
	//	val, err := strconv.Atoi(strVal)
	//	if err != nil || val <= 0 {
	//		log.Fatalf("invalid Flags `%s` at line %d", strVal, o.Position.Line)
	//	}
	//	p.Flags = uint32(val)
	//} else if strings.Index(o.Name, "ext.Url") >= 0 {
	//	p.Url = strings.TrimSpace(o.Constant.Source)
	//} else if strings.Index(o.Name, "ext.UserType") >= 0 {
	//	p.UserType = strings.TrimSpace(o.Constant.Source)
	//} else if strings.Index(o.Name, "ext.Perm") >= 0 {
	//	p.Perm = strings.TrimSpace(o.Constant.Source)
	//}
	//
	//if p.options == nil {
	//	p.options = map[string]string{}
	//}
	//
	//p.options[o.Name] = o.Constant.Source
}

func (p *ProtoVisitor) VisitImport(i *proto.Import) {
}

func (p *ProtoVisitor) VisitNormalField(i *proto.NormalField) {
	if p.Message == nil {
		return
	}

	p.Message.FieldList = append(p.Message.FieldList, &MessageField{NormalField: i})
}

func (p *ProtoVisitor) VisitEnumField(i *proto.EnumField) {
	p.EnumFieldList = append(p.EnumFieldList, i)
	//
	//if CurrentMod != "" {
	//	AllErrCodes.Set(CurrentMod, i.Name, uint32(i.Integer))
	//
	//	if i.InlineComment != nil && len(i.InlineComment.Lines) > 0 {
	//		line := strings.TrimSpace(i.InlineComment.Lines[0])
	//		if line != "" {
	//			AllErrCodes.SetMsg(CurrentMod, i.Name, line)
	//		}
	//	}
	//}
}

func (p *ProtoVisitor) VisitEnum(e *proto.Enum) {
}

func (p *ProtoVisitor) VisitComment(e *proto.Comment) {
}

func (p *ProtoVisitor) VisitOneof(o *proto.Oneof) {
}

func (p *ProtoVisitor) VisitOneofField(o *proto.OneOfField) {
}

func (p *ProtoVisitor) VisitReserved(rs *proto.Reserved) {
}

func (p *ProtoVisitor) VisitRPC(rpc *proto.RPC) {
	p.Name = rpc.Name
	//p.RequestType = rpc.RequestType
	//p.ReturnsType = rpc.RequestType
	//p.StreamsRequest = rpc.StreamsRequest
}

func (p *ProtoVisitor) VisitMapField(f *proto.MapField) {
	if p.Message == nil {
		return
	}

	p.Message.FieldList = append(p.Message.FieldList, &MessageField{MapField: f})
}

func (p *ProtoVisitor) VisitGroup(g *proto.Group) {
}
func (p *ProtoVisitor) VisitExtensions(e *proto.Extensions) {
}
