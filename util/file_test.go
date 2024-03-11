package util

import (
	"os"
	"reflect"
	"testing"
)

func TestIsFileExist(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		// TODO: Add test cases.
		{"test1", args{"./file.go"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsFileExist(tt.args.name); got != tt.want {
				t.Errorf("IsFileExist() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpenFile(t *testing.T) {
	type args struct {
		fPath string
	}
	tests := []struct {
		name    string
		args    args
		want    *os.File
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := OpenFile(tt.args.fPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("OpenFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("OpenFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
