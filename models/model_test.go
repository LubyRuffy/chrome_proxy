package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResult_Bytes(t *testing.T) {
	type fields struct {
		Code     int
		Message  string
		Url      string
		Data     string
		Title    string
		Location string
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			name: "测试 Bytes 函数",
			fields: fields{
				Code:     200,
				Message:  "success",
				Url:      "https://example.com",
				Data:     "sample",
				Title:    "this is a test result",
				Location: "https://example.com",
			},
			want: []byte("{\"code\":200,\"message\":\"success\",\"url\":\"https://example.com\",\"data\":\"sample\"," +
				"\"title\":\"this is a test result\",\"location\":\"https://example.com\"}"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Result{
				Code:     tt.fields.Code,
				Message:  tt.fields.Message,
				Url:      tt.fields.Url,
				Data:     tt.fields.Data,
				Title:    tt.fields.Title,
				Location: tt.fields.Location,
			}
			assert.Equalf(t, tt.want, r.Bytes(), "Bytes()")
		})
	}
}
