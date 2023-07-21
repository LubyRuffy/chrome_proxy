package main

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

func Test_chromeActions(t *testing.T) {
	buf := []byte{}
	type args struct {
		in      ChromeActionInput
		logf    func(string, ...interface{})
		timeout int
		actions []chromedp.Action
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试正常screen picture",
			args: args{
				in: ChromeActionInput{
					URL: "https://www.baidu.com",
				},
				logf:    func(s string, i ...interface{}) {},
				timeout: 10,
				actions: []chromedp.Action{
					chromedp.Sleep(time.Second * time.Duration(1)),
					chromedp.CaptureScreenshot(&buf),
				},
			},
		},
		{
			name: "测试钓鱼页面proxy & UA",
			args: args{
				in: ChromeActionInput{
					URL:       "http://shop.bnuzac.com/articles.php/about/456415?newsid=oaxn1d.html",
					Proxy:     "socks5://127.0.0.1:7890",
					UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
				},
				logf:    func(s string, i ...interface{}) {},
				timeout: 10,
				actions: []chromedp.Action{
					chromedp.Sleep(time.Second * time.Duration(1)),
					chromedp.CaptureScreenshot(&buf),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := chromeActions(tt.args.in, func(s string, i ...interface{}) {}, tt.args.timeout, tt.args.actions...)
			assert.Nil(t, err)
			assert.NotNil(t, buf)
			t.Logf("screenshot result: %s", buf[0:5])
		})
	}
}

func TestResult_Bytes(t *testing.T) {
	debug = true
	data, err := RenderDom(&chromeParam{
		ChromeActionInput: ChromeActionInput{
			URL: "https://bgp.he.net/ip/106.75.29.24",
		},
		Sleep:   5,
		Timeout: 30,
	})
	assert.Nil(t, err)
	assert.Contains(t, data.html, "wrzxfw.top")
}

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
			err := chromedp.Run(ctx, fullScreenshot(tt.args.url, 90, &buf))
			assert.Nil(t, err)

			gotResult, err := AddUrlToTitle(tt.args.url, buf, tt.args.useTimeStamp)
			assert.Nil(t, err)
			assert.Greater(t, len(gotResult), len(buf))

			// 效果展示
			var fn string
			fn, err = WriteTempFile(".png", func(f *os.File) error {
				_, err = f.Write(gotResult)
				return err
			})
			log.Printf("save modified pic into: %s", fn)
		})
	}
}

func TestAddUrlToNavBar(t *testing.T) {
	type args struct {
		url          string
		useTimeStamp bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试带有导航栏的截图",
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
			var title string
			err := chromedp.Run(ctx, fullScreenshot(tt.args.url, 90, &buf), chromedp.Title(&title))
			assert.Nil(t, err)

			gotResult, err := AddUrlNavBar(tt.args.url, title, buf, tt.args.useTimeStamp)
			assert.Nil(t, err)
			assert.Greater(t, len(gotResult), len(buf))

			// 效果展示
			var fn string
			fn, err = WriteTempFile(".png", func(f *os.File) error {
				_, err = f.Write(gotResult)
				return err
			})
			log.Printf("save modified pic into: %s", fn)
		})
	}
}

func Test_renderURLDOM(t *testing.T) {
	type args struct {
		in      chromeParam
		logf    func(string, ...interface{})
		timeout int
		actions []chromedp.Action
	}
	tests := []struct {
		name string
		args args
		want *RenderDomOutput
	}{
		{
			name: "测试正常 render dom",
			args: args{
				in: chromeParam{
					Sleep:        5,
					Timeout:      30,
					AddUrl:       false,
					AddTimeStamp: false,
					ChromeActionInput: ChromeActionInput{
						URL: "https://www.baidu.com",
					},
				},
				logf: func(s string, i ...interface{}) {},
			},
			want: &RenderDomOutput{
				html:     "百度",
				title:    "百度一下，你就知道",
				location: "https://www.baidu.com/",
			},
		},
		{
			name: "测试自定义proxy & UA",
			args: args{
				in: chromeParam{
					Sleep:        5,
					Timeout:      30,
					AddUrl:       false,
					AddTimeStamp: false,
					ChromeActionInput: ChromeActionInput{
						URL:       "https://www.fofa.info",
						Proxy:     "socks5://127.0.0.1:7890",
						UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
					},
				},
				logf: func(s string, i ...interface{}) {},
			},
			want: &RenderDomOutput{
				html:     "FOFA",
				title:    "FOFA Search Engine",
				location: "fofa.info",
			},
		},
		{
			name: "钓鱼测试 proxy & UA",
			args: args{
				in: chromeParam{
					Sleep:        5,
					Timeout:      30,
					AddUrl:       false,
					AddTimeStamp: false,
					ChromeActionInput: ChromeActionInput{
						URL:       "http://asd.naeuib12123d.xyz/a.html#/",
						UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
						Proxy:     "socks5://127.0.0.1:7890",
					},
				},
				logf: func(s string, i ...interface{}) {},
			},
			want: &RenderDomOutput{
				html:     "ETC联网升级",
				title:    "认证中心",
				location: "http://asd.naeuib12123d.xyz/a.html#/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := RenderDom(&tt.args.in)
			assert.Nil(t, err)
			assert.Contains(t, out.html, tt.want.html)
			assert.Contains(t, out.location, tt.want.location)
			assert.Equal(t, out.title, tt.want.title)
			t.Logf("RenderDom result: %s", out.html[0:5])
		})
	}
}

func TestDrawNavInfo(t *testing.T) {
	type args struct {
		url          string
		useTimeStamp bool
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "测试带有导航栏的截图",
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
			var title string
			err := chromedp.Run(ctx, fullScreenshot(tt.args.url, 90, &buf), chromedp.Title(&title))
			assert.Nil(t, err)

			gotResult, err := drawNavInfo(tt.args.url, title)
			assert.Nil(t, err)
			assert.Greater(t, len(gotResult), 0)

			log.Printf("save modified pic into: %s", gotResult)
		})
	}
}

func Test_gptProcess(t *testing.T) {
	type args struct {
		in ChromeActionInput
	}
	tests := []struct {
		name    string
		args    args
		wantRes string
	}{
		{
			name: "测试 GPT 验证 - 成功",
			args: args{in: ChromeActionInput{
				URL: "http://192.248.178.124:3002/",
			}},
			wantRes: "Hello",
		},
		{
			name: "测试 GPT 验证 - 失败",
			args: args{in: ChromeActionInput{
				URL: "http://119.135.192.236:8002/",
			}},
			wantRes: "ChatGPT error 429",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := gptProcess(tt.args.in)
			assert.Truef(t, strings.Contains(got, tt.wantRes), "%s = gptProcess(%v) != %s", got, tt.args.in.URL, tt.wantRes)
		})
	}
}
