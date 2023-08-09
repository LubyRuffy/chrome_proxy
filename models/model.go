package models

import (
	"encoding/json"
)

var (
	// DefaultUserAgent 默认 UA
	DefaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.124 Safari/537.36 Edg/102.0.1245.44"

	// Debug 是否为调试模式
	Debug = false

	// DefaultTmpFilePrefix 默认前缀
	DefaultTmpFilePrefix = "chrome_proxy_"
)

// ChromeActionInput chrome 渲染输入字段
type ChromeActionInput struct {
	URL       string `json:"url"`
	Proxy     string `json:"proxy,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	Sleep     int    `json:"sleep"`
	Timeout   int    `json:"timeout"`
}

// ChromeParam Chrome 渲染输入字段
type ChromeParam struct {
	AddUrl       bool `json:"add_url"` // 在截图中展示url地址
	AddTimeStamp bool `json:"add_time_stamp"`
	ChromeActionInput
}

// RenderDomOutput Dom 渲染输出结果
type RenderDomOutput struct {
	Html     string
	Title    string
	Location string
}

// ScreenshotOutput 截图输出内容
type ScreenshotOutput struct {
	Data     []byte
	Title    string
	Location string
}

// Result 统一输出结果
type Result struct {
	Code     int    `json:"code"`
	Message  string `json:"message,omitempt"`
	Url      string `json:"url,omitempty"`
	Data     string `json:"data,omitempty"`
	Title    string `json:"title,omitempty"`
	Location string `json:"location,omitempty"`
}

func (r Result) Bytes() []byte {
	d, _ := json.Marshal(r)
	return d
}
