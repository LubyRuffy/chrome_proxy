package render_dom

import (
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/chromedp/chromedp"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_renderURLDOM(t *testing.T) {
	type args struct {
		in      *models.ChromeParam
		logf    func(string, ...interface{})
		timeout int
		actions []chromedp.Action
	}
	tests := []struct {
		name string
		args args
		want *models.RenderDomOutput
	}{
		{
			name: "测试正常 render dom",
			args: args{
				in: &models.ChromeParam{
					AddUrl:       false,
					AddTimeStamp: false,
					ChromeActionInput: models.ChromeActionInput{
						URL:     "https://www.baidu.com",
						Sleep:   5,
						Timeout: 30,
					},
				},
				logf: func(s string, i ...interface{}) {},
			},
			want: &models.RenderDomOutput{
				Html:     "百度",
				Title:    "百度一下，你就知道",
				Location: "https://www.baidu.com/",
			},
		},
		{
			name: "测试自定义proxy & UA",
			args: args{
				in: &models.ChromeParam{
					AddUrl:       false,
					AddTimeStamp: false,
					ChromeActionInput: models.ChromeActionInput{
						URL:       "https://www.fofa.info",
						Proxy:     "socks5://127.0.0.1:7890",
						UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
						Sleep:     5,
						Timeout:   30,
					},
				},
				logf: func(s string, i ...interface{}) {},
			},
			want: &models.RenderDomOutput{
				Html:     "FOFA",
				Title:    "FOFA Search Engine",
				Location: "fofa.info",
			},
		},
		{
			name: "钓鱼测试 proxy & UA",
			args: args{
				in: &models.ChromeParam{
					AddUrl:       false,
					AddTimeStamp: false,
					ChromeActionInput: models.ChromeActionInput{
						URL:       "http://asd.naeuib12123d.xyz/a.html#/",
						UserAgent: "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/86.0.4240.198 Safari/537.36",
						Proxy:     "socks5://127.0.0.1:7890",
						Sleep:     5,
						Timeout:   30,
					},
				},
				logf: func(s string, i ...interface{}) {},
			},
			want: &models.RenderDomOutput{
				Html:     "請登陸用戶後臺綁定或啟動網站", // "ETC联网升级",
				Title:    "",               // "认证中心",
				Location: "http://asd.naeuib12123d.xyz/a.html#/",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := RenderDom(tt.args.in)
			assert.Nil(t, err)
			assert.Contains(t, out.Html, tt.want.Html)
			assert.Contains(t, out.Location, tt.want.Location)
			assert.Equal(t, out.Title, tt.want.Title)
			t.Logf("RenderDom result: %s", out.Html[0:5])
		})
	}
}
