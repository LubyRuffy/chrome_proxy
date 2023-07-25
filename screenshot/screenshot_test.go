package screenshot

import (
	"context"
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
