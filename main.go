package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/chromedp"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

var (
	defaultUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.124 Safari/537.36 Edg/102.0.1245.44"
	debug                = false
	defaultTmpFilePrefix = "chrome_proxy_"
)

type ScreenshotOutput struct {
	Data     []byte
	Title    string
	Location string
}

type ChromeActionInput struct {
	URL       string `json:"url"`
	Proxy     string `json:"proxy,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}

// chromeActions 完成chrome的headless操作
func chromeActions(in ChromeActionInput, logf func(string, ...interface{}), timeout int, actions ...chromedp.Action) error {
	var err error

	// set user-agent
	if in.UserAgent == "" {
		in.UserAgent = defaultUserAgent
	}

	// prepare the chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("incognito", true), // 隐身模式
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.Flag("enable-automation", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.IgnoreCertErrors,
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.NoSandbox,
		chromedp.DisableGPU,
		chromedp.UserAgent(in.UserAgent), // chromedp.Flag("user-agent", defaultUserAgent)
		chromedp.WindowSize(1024, 768),
	)

	// set proxy if exists
	if in.Proxy != "" {
		opts = append(opts, chromedp.ProxyServer(in.Proxy))
	}

	if debug {
		opts = append(chromedp.DefaultExecAllocatorOptions[:2],
			chromedp.DefaultExecAllocatorOptions[3:]...)
		opts = append(opts, chromedp.Flag("auto-open-devtools-for-tabs", true))
	}

	allocCtx, bcancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer func() {
		bcancel()
		b := chromedp.FromContext(allocCtx).Browser
		if b != nil && b.Process() != nil {
			b.Process().Signal(os.Kill)
		}
	}()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logf))
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	realActions := []chromedp.Action{
		chromedp.ActionFunc(func(cxt context.Context) error {
			// 等待完成，要么是body出来了，要么是资源加载完成
			ch := make(chan error, 1)
			go func(eCh chan error) {
				err := chromedp.Navigate(in.URL).Do(cxt)
				if err != nil {
					eCh <- err
				}
				var htmlDom string
				err = chromedp.WaitReady("body", chromedp.ByQuery).Do(cxt)
				if err == nil {
					if err2 := chromedp.OuterHTML("html", &htmlDom).Do(cxt); err != nil {
						log.Println("[DEBUG] fetch html failed:", err2)
					}
				}
				// 20211219发现如果存在JS前端框架 (如vue, react...) 执行等待读取.
				html2Low := strings.ToLower(htmlDom)
				if strings.Contains(html2Low, "javascript") || strings.Contains(html2Low, "</script>'") {
					err2 := chromedp.WaitVisible("div", chromedp.ByQuery).Do(cxt)
					if err2 = chromedp.OuterHTML("html", &htmlDom).Do(cxt); err2 != nil {
						// extra error, doesnt affect anything else
						log.Println("[DEBUG] fetch html failed:", err2)
					}
				}

				eCh <- err
			}(ch)

			select {
			case <-time.After(time.Duration(timeout) * time.Second):
			case err := <-ch:
				if err != nil {
					return err
				}
			}

			return nil
		}),
	}

	realActions = append(realActions, actions...)

	// run task list
	err = chromedp.Run(ctx, realActions...)

	return err
}

type chromeParam struct {
	Sleep        int  `json:"sleep"`
	Timeout      int  `json:"timeout"`
	AddUrl       bool `json:"add_url"` // 在截图中展示url地址
	AddTimeStamp bool `json:"add_time_stamp"`
	ChromeActionInput
}

func screenshotURL(options *chromeParam) (*ScreenshotOutput, error) {
	log.Println("screenshot of url:", options.URL)

	var buf []byte
	var actions []chromedp.Action
	if options.Sleep > 0 {
		actions = append(actions, chromedp.Sleep(time.Second*time.Duration(options.Sleep)))
	}
	actions = append(actions, chromedp.CaptureScreenshot(&buf))

	var title string
	actions = append(actions, chromedp.Title(&title))
	var url string
	actions = append(actions, chromedp.Location(&url))

	err := chromeActions(options.ChromeActionInput, func(s string, i ...interface{}) {

	}, options.Timeout, actions...)
	if err != nil {
		return nil, fmt.Errorf("screenShot failed(%w): %s", err, options.URL)
	}

	log.Printf("finished screenshot for %s", options.URL)

	// 在截图中添加当前请求地址
	if options.AddUrl {
		tmp, err := AddUrlToTitle(options.URL, buf, options.AddTimeStamp)
		if err != nil {
			return nil, fmt.Errorf("add url title failed(%w): %s", err, options.URL)
		}
		buf = tmp
	}

	return &ScreenshotOutput{
		Data:     buf,
		Title:    title,
		Location: url,
	}, err
}

type renderDomOutput struct {
	html     string
	title    string
	location string
}

// renderURLDOM 生成单个url的domhtml
func renderURLDOM(options *chromeParam) (*renderDomOutput, error) {
	log.Println("renderURLDOM of url:", options.URL)

	var html string
	var actions []chromedp.Action
	if options.Sleep > 0 {
		actions = append(actions, chromedp.Sleep(time.Second*time.Duration(options.Sleep)))
	}

	actions = append(actions, chromedp.ActionFunc(func(ctx context.Context) error {
		node, err := dom.GetDocument().Do(ctx)
		if err != nil {
			return err
		}
		html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
		return err
	}))

	var title string
	actions = append(actions, chromedp.Title(&title))
	var location string
	actions = append(actions, chromedp.Location(&location))

	err := chromeActions(options.ChromeActionInput, func(s string, i ...interface{}) {

	}, options.Timeout, actions...)

	if err != nil {
		return nil, fmt.Errorf("renderURLDOM failed(%w): %s", err, options.URL)
	}

	return &renderDomOutput{
		html:     html,
		title:    title,
		location: location,
	}, err
}

func getOptionFromRequest(r *http.Request) (*chromeParam, error) {
	var options chromeParam
	err := json.NewDecoder(r.Body).Decode(&options)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if options.Timeout == 0 {
		options.Timeout = 20
	}
	return &options, nil

}

type result struct {
	Code     int    `json:"code"`
	Message  string `json:"message,omitempt"`
	Url      string `json:"url,omitempty"`
	Data     string `json:"data,omitempty"`
	Title    string `json:"title,omitempty"`
	Location string `json:"location,omitempty"`
}

func (r result) Bytes() []byte {
	d, _ := json.Marshal(r)
	return d
}

func main() {
	addr := flag.String("addr", ":5558", "http server listen address")
	flag.Parse()

	http.HandleFunc("/screenshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		options, err := getOptionFromRequest(r)
		if options == nil {
			w.Write(result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		screenshotResult, err := screenshotURL(options)
		if err != nil {
			w.Write(result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(result{
			Code:     200,
			Url:      options.URL,
			Data:     base64.StdEncoding.EncodeToString(screenshotResult.Data),
			Title:    screenshotResult.Title,
			Location: screenshotResult.Location,
		}.Bytes())
	})

	http.HandleFunc("/renderDom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		options, err := getOptionFromRequest(r)
		if options == nil {
			w.Write(result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		data, err := renderURLDOM(options)
		if err != nil {
			w.Write(result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(result{
			Code:     200,
			Url:      options.URL,
			Data:     data.html,
			Title:    data.title,
			Location: data.location,
		}.Bytes())
	})

	log.Println("listen at address:", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

// AddUrlToTitle 通过html转换对整个screenshot截图结果进行处理，添加标题栏并在其中写入访问的url地址
func AddUrlToTitle(url string, picBuf []byte, hasTimeStamp bool) (result []byte, err error) {
	htmlPart1 := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Title</title>
    <style>
        .window {
                border-radius: 5px;
                -moz-box-shadow:1em 1em 3em #333333; -webkit-box-shadow:1em 1em 3em #333333; box-shadow:1em 1em 3em #333333;
                margin: 25px;
            }
        .window-header .btn {
            width: 10px;
            height: 10px;
            margin: 6px 0 6px 10px;
            border-radius: 50%;
            padding: 0;
            display: inline-block;
            font-size: 14px;
            font-weight: 400;
            line-height: 1.42857143;
            text-align: center;
            white-space: nowrap;
            vertical-align: middle;
            touch-action: manipulation;
            cursor: pointer;
        }
        .window-header .btn.red {
            border: 1px solid #ff3125;
            background-color: #ff6158;
        }
        .window-header .btn.yellow {
            border: 1px solid #f9ab00;
            background-color: #ffbd2d;
        }
        .window-header .btn.green {
            border: 1px solid #21a435;
            background-color: #2ace43;
        }
        .window-header {
            display: block;
            border-radius: 5px 5px 0 0;
            border-top: solid 1px #f3f1f3;
            background-image: -webkit-linear-gradient(#e3dfe3,#d0cdd0);
            background-image: linear-gradient(#e3dfe3,#d0cdd0);
            width: 100%;
            height: 22px;
        }
        body {
            font-family: "Helvetica Neue",Helvetica, "microsoft yahei", arial, STHeiTi, sans-serif;
        }
    </style>
</head>
<body>
    <div>
        <div class="window">
            <div class="window-header">
                <div class="btn red"></div>
                <div class="btn yellow"></div>
                <div class="btn green"></div>
`
	htmlTitle := `<div class="btn" style="margin-top: -7px;margin-left: 1%;">
                    <b style="color:#48576a">`
	htmlTimeStamp := `</b>
                </div>
                <div class="btn" style="margin-top: 2px;margin-right: 18%;float: right;">
                    <b style="color:#48576a">`
	htmlBase64 := `</b>
                </div>
            </div>
            <div style="max-height:800px;overflow:hidden;">
                <img  style="width:100%;" src="data:image/png;base64,`
	htmlPart3 := `" />
            </div>
        </div>
    </div>
</body>
</html>`

	// 生成的图片通过base64加密
	encodedBase64 := base64.StdEncoding.EncodeToString(picBuf)

	// 合成新的html文件
	html := append(append([]byte(htmlPart1), []byte(htmlTitle)...), []byte(url)...)

	// 添加时间戳
	if hasTimeStamp {
		curTime := time.Now().Format(`2006-01-02 15:04:05`)
		html = append(html, []byte(htmlTimeStamp)...)
		html = append(html, []byte(curTime)...)
	}

	// 添加
	html = append(append(append(html, []byte(htmlBase64)...), []byte(encodedBase64)...), []byte(htmlPart3)...)
	var fn string
	fn, err = WriteTempFile(".html", func(f *os.File) error {
		_, err = f.Write(html)
		return err
	})

	// 将html文件进行截图
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	var buf []byte
	if err = chromedp.Run(ctx, fullScreenshot(`file://`+fn, 100, &buf)); err != nil {
		return nil, err
	}

	return buf, err
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func fullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.FullScreenshot(res, quality),
	}
}

// WriteTempFile 写入临时文件
// 如果writeF是nil，就只返回生成的一个临时空文件路径
// 返回文件名和错误
func WriteTempFile(ext string, writeF func(f *os.File) error) (fn string, err error) {
	var f *os.File
	if len(ext) > 0 {
		ext = "*" + ext
	}
	f, err = os.CreateTemp(os.TempDir(), defaultTmpFilePrefix+ext)
	if err != nil {
		return
	}
	defer f.Close()

	fn = f.Name()

	if writeF != nil {
		err = writeF(f)
		if err != nil {
			return
		}
	}
	return
}
