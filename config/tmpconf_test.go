package config

import "testing"

func TestJoinGoPackage(t *testing.T) {
	type args struct {
		serviceGroup string
		protoName    string
	}
	tests := []struct {
		name    string
		prepare func()
		args    args
		want    string
	}{
		{
			"test1",
			func() {
				GlobalConfig.GoModulePrefix = "github.com"
				GlobalConfig.TemplatesConf.Template.RelativeDir.Client = []string{"pkg", "{SERVICE_NAME"}
			},
			args{"ml444", "user"},
			"github.com/ml444/pkg/user",
		},
		{
			"test2",
			func() {
				GlobalConfig.GoModulePrefix = "github.com"
				GlobalConfig.TemplatesConf.Template.RelativeDir.Client = []string{"pkg", "{SERVICE_NAME"}
			},
			args{"", "user"},
			"github.com/pkg/user",
		},
		{
			"test3",
			func() {
				GlobalConfig.GoModulePrefix = ""
				GlobalConfig.TemplatesConf.Template.RelativeDir.Client = []string{"pkg", "{SERVICE_NAME"}

			},
			args{"", "user"},
			"pkg/user",
		},
		{
			"test4",
			func() {
				GlobalConfig.GoModulePrefix = ""
				GlobalConfig.TemplatesConf.Template.RelativeDir.Client = []string{"pkg", "{SERVICE_NAME"}
			},
			args{"", "./base/pkg/user.proto"},
			"base/pkg/user",
		},
		{
			"test5",
			func() {
				GlobalConfig.GoModulePrefix = ""
				GlobalConfig.TemplatesConf.Template.RelativeDir.Client = []string{"pkg", "{SERVICE_NAME"}
			},
			args{"", "user.proto"},
			"pkg/user",
		},
		{
			"test6",
			func() {
				GlobalConfig.GoModulePrefix = "github.com"
				GlobalConfig.TemplatesConf.Template.RelativeDir.Client = []string{"pkg", "{SERVICE_NAME"}
			},
			args{"", "./base/pkg1/user.proto"},
			"github.com/base/pkg1/user",
		},
	}
	for _, tt := range tests {
		_ = InitGlobalVar()
		t.Run(tt.name, func(t *testing.T) {
			tt.prepare()
			if got := JoinGoPackage(tt.args.serviceGroup, tt.args.protoName); got != tt.want {
				t.Errorf("JoinGoPackage() = %v, want %v", got, tt.want)
			}
		})
	}
}
