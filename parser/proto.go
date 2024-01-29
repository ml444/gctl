package parser

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
	log "github.com/ml444/glog"
)

// ParseProtoFile 解析proto文件
func ParseProtoFile(protoFilepath string) (*CtxData, error) {
	reader, err := os.Open(protoFilepath)
	if err != nil {
		log.Errorf("err: %v", err)
		return nil, err
	}
	defer func() {
		err = reader.Close()
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
	ctxData := NewCtxData()
	ctxData.FilePath = protoFilepath
	handlePackage := func(p *proto.Package) {
		ctxData.PackageName = p.Name
	}

	handleOptions := func(o *proto.Option) {
		if ctxData.Options == nil {
			ctxData.Options = make(map[string]string)
		}
		ctxData.Options[o.Name] = o.Constant.Source
	}

	handleService := func(s *proto.Service) {
		svc := &Service{
			ServiceName: s.Name,
			//RpcList:     nil,
		}

		ctxData.ServiceList = append(ctxData.ServiceList, svc)
		log.Info("===> serviceName:", s.Name)
	}

	handleRpcMethod := func(m *proto.RPC) {
		options := map[string]string{}
		log.Infof("=====> rpcName: %s", m.Name)
		method := &RpcMethod{
			Name:         m.Name,
			RequestType:  m.RequestType,
			ResponseType: m.ReturnsType,
			Options:      options,
		}

		if m.Comment != nil {
			method.CommentLines = m.Comment.Lines
		}

		parent := &ProtoVisitor{}
		m.Parent.Accept(parent)
		for _, svc := range ctxData.ServiceList {
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
			ctxData.ErrCodeList = append(
				ctxData.ErrCodeList,
				Enum{Name: e.Name, EnumFieldList: enum.EnumFieldList})
		} else if strings.HasPrefix(e.Name, "ListOpt") {
			if ctxData.ListOptionMap == nil {
				ctxData.ListOptionMap = map[string]*ListReqOption{}
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
				if ctxData.ListOptionMap == nil {
					ctxData.ListOptionMap = map[string]*ListReqOption{
						reqName: &listReqOpt,
					}
				} else {
					ctxData.ListOptionMap[reqName] = &listReqOpt
				}
			}
		}
	}

	handleImport := func(i *proto.Import) {
		if ctxData.ImportMap == nil {
			ctxData.ImportMap = make(map[string]bool)
		}
		if ctxData.ImportMap[i.Filename] {
			return
		}
		defer func() {
			ctxData.ImportMap[i.Filename] = true
		}()

		if strings.HasPrefix(i.Filename, "google/") {
			return
		}
	}

	handleMessage := func(m *proto.Message) {
		Msg := &Message{Name: m.Name}
		vv := &ProtoVisitor{Message: Msg}
		for _, v := range m.Elements {
			v.Accept(vv)
		}

		if strings.HasPrefix(m.Name, "Model") {
			ctxData.ModelList = append(ctxData.ModelList, m)
			ctxData.ModelFieldList = append(ctxData.ModelFieldList, vv.FieldNameList...)
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
	if ctxData.PackageName == "" {
		_, filename := filepath.Split(ctxData.FilePath)
		ctxData.PackageName = strings.TrimSuffix(filename, ".proto")
	}

	// unique ModelFieldList
	ctxData.ModelFieldList = ToUnique(ctxData.ModelFieldList)
	return ctxData, nil
}

func ToUnique(list []string) []string {
	var m = make(map[string]struct{})
	for _, v := range list {
		m[v] = struct{}{}
	}
	var newList []string
	for k := range m {
		newList = append(newList, k)
	}
	return newList
}

type ProtoVisitor struct {
	//proto.Visitor
	//proto.RPC
	Name          string
	FieldNameList []string
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
}

func (p *ProtoVisitor) VisitImport(i *proto.Import) {
}

func (p *ProtoVisitor) VisitNormalField(i *proto.NormalField) {
	if p.Message == nil {
		return
	}
	p.FieldNameList = append(p.FieldNameList, i.Name)
	p.Message.FieldList = append(p.Message.FieldList, &MessageField{NormalField: i})
}

func (p *ProtoVisitor) VisitEnumField(i *proto.EnumField) {
	p.EnumFieldList = append(p.EnumFieldList, i)
}

func (p *ProtoVisitor) VisitMapField(f *proto.MapField) {
	if p.Message == nil {
		return
	}

	p.Message.FieldList = append(p.Message.FieldList, &MessageField{MapField: f})
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

func (p *ProtoVisitor) VisitGroup(g *proto.Group) {
}
func (p *ProtoVisitor) VisitExtensions(e *proto.Extensions) {
}
