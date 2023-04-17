package cmd

import "testing"

func TestGetStatusCodeFromComment(t *testing.T) {
	type args struct {
		comment string
	}
	tests := []struct {
		name    string
		args    args
		want    int32
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name:    "",
			args:    args{comment: "@status_code: 400"},
			want:    400,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetStatusCodeFromComment(tt.args.comment)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetStatusCodeFromComment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetStatusCodeFromComment() got = %v, want %v", got, tt.want)
			}
		})
	}
}
