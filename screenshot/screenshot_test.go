package screenshot

import (
	"context"
	"fmt"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/LubyRuffy/chrome_proxy/utils"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

func TestAddUrlToTitle(t *testing.T) {
	type args struct {
		url          string
		useTimeStamp bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试截图添加url地址",
			args: args{url: `https://fofa.info`, useTimeStamp: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := chromedp.NewContext(
				context.Background(),
			)
			defer cancel()

			var buf []byte
			err := chromedp.Run(ctx, FullScreenshot(tt.args.url, 90, &buf))
			assert.Nil(t, err)

			gotResult, err := AddUrlToTitle(tt.args.url, buf, tt.args.useTimeStamp)
			assert.Nil(t, err)
			assert.Greater(t, len(gotResult), len(buf))

			// 效果展示
			var fn string
			fn, err = utils.WriteTempFile(".png", func(f *os.File) error {
				_, err = f.Write(gotResult)
				return err
			})
			log.Printf("save modified pic into: %s", fn)
		})
	}
}

func TestScreenshotURL(t *testing.T) {
	type args struct {
		options *models.ChromeParam
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "正常截图",
			args: args{options: &models.ChromeParam{
				AddUrl:       false,
				AddTimeStamp: false,
				ChromeActionInput: models.ChromeActionInput{
					URL:     "https://fofa.info",
					Sleep:   2,
					Timeout: 30,
				},
			}},
			wantErr: assert.NoError,
		},
		{
			name: "带base64输出",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ScreenshotURL(tt.args.options)
			if !tt.wantErr(t, err, fmt.Sprintf("ScreenshotURL(%v)", tt.args.options)) {
				return
			}
			fmt.Printf("pic data %s", got.Data[:5])
		})
	}
}
