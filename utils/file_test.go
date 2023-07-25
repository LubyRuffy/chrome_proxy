package utils

import (
	"os"
	"strings"
	"testing"
)

func TestWriteTempFile(t *testing.T) {
	type args struct {
		ext    string
		writeF func(f *os.File) error
	}
	tests := []struct {
		name    string
		args    args
		wantFn  string
		wantErr bool
	}{
		{
			name: "测试文件写入",
			args: args{
				ext: ".png",
				writeF: func(f *os.File) error {
					_, err := f.Write([]byte("sample"))
					return err
				},
			},
			wantFn:  "chrome_proxy_",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFn, err := WriteTempFile(tt.args.ext, tt.args.writeF)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteTempFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !strings.Contains(gotFn, tt.wantFn) {
				t.Errorf("WriteTempFile() gotFn = %v, want %v", gotFn, tt.wantFn)
			}
		})
	}
}
