package parser

import (
	"testing"
)

func Test_stuct2map(t *testing.T) {
	pd := &ProtoData{
		FilePath:    "/home/user",
		PackageName: "testUser",
	}
	m, err := Struct2map(pd)
	if err != nil {
		t.Error(err)
	}
	t.Log(m)
	if v, ok := m["FilePath"]; !ok {
		t.Error("not found")
	} else {
		t.Log(v)
	}
}

func TestParseProto(t *testing.T) {
	protoFile := "../user.proto"
	pd, _ := ParseProto(protoFile)
	t.Log(pd)
}
