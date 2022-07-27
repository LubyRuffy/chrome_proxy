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
	"strings"
	"time"
)

var (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/102.0.5005.124 Safari/537.36 Edg/102.0.1245.44"
)

//chromeActions 完成chrome的headless操作
func chromeActions(u string, logf func(string, ...interface{}), timeout int, actions ...chromedp.Action) error {
	var err error
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
		chromedp.UserAgent(defaultUserAgent), // chromedp.Flag("user-agent", defaultUserAgent)
		chromedp.WindowSize(1024, 768),
	)

	allocCtx, bcancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer bcancel()

	ctx, cancel := chromedp.NewContext(allocCtx, chromedp.WithLogf(logf))
	ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	realActions := []chromedp.Action{
		chromedp.ActionFunc(func(cxt context.Context) error {
			// 等待完成，要么是body出来了，要么是资源加载完成
			ch := make(chan error, 1)
			go func(eCh chan error) {
				err := chromedp.Navigate(u).Do(cxt)
				if err != nil {
					eCh <- err
				}
				var htmlDom string
				err = chromedp.WaitReady("body", chromedp.ByQuery).Do(cxt)
				if err == nil {
					if err := chromedp.OuterHTML("html", &htmlDom).Do(cxt); err != nil {
						log.Println("[DEBUG] fetch html failed:", err)
					}
				}
				// 20211219发现如果存在JS前端框架 (如vue, react...) 执行等待读取.
				html2Low := strings.ToLower(htmlDom)
				if strings.Contains(html2Low, "javascript") || strings.Contains(html2Low, "</script>'") {
					err = chromedp.WaitVisible("div", chromedp.ByQuery).Do(cxt)
					if err := chromedp.OuterHTML("html", &htmlDom).Do(cxt); err != nil {
						log.Println("[DEBUG] fetch html failed:", err)
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
	URL     string
	Sleep   int
	Timeout int
}

func screenshotURL(options *chromeParam) ([]byte, error) {
	log.Println("screenshot of url:", options.URL)

	var buf []byte
	var actions []chromedp.Action
	if options.Sleep > 0 {
		actions = append(actions, chromedp.Sleep(time.Second*time.Duration(options.Sleep)))
	}
	actions = append(actions, chromedp.CaptureScreenshot(&buf))

	err := chromeActions(options.URL, func(s string, i ...interface{}) {

	}, options.Timeout, actions...)
	if err != nil {
		return nil, fmt.Errorf("screenShot failed(%w): %s", err, options.URL)
	}

	return buf, nil
}

// renderURLDOM 生成单个url的domhtml
func renderURLDOM(options *chromeParam) (string, error) {
	log.Println("renderURLDOM of url:", options.URL)

	var html string
	err := chromeActions(options.URL, func(s string, i ...interface{}) {

	}, options.Timeout,
		chromedp.ActionFunc(func(ctx context.Context) error {
			node, err := dom.GetDocument().Do(ctx)
			if err != nil {
				return err
			}
			html, err = dom.GetOuterHTML().WithNodeID(node.NodeID).Do(ctx)
			return err
		}))
	if err != nil {
		return "", fmt.Errorf("renderURLDOM failed(%w): %s", err, options.URL)
	}

	return html, err
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
	Code    int    `json:"code"`
	Message string `json:"message,omitempt"`
	Url     string `json:"url,omitempty"`
	Data    string `json:"data,omitempty"`
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

		data, err := screenshotURL(options)
		if err != nil {
			w.Write(result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(result{
			Code: 200,
			Url:  options.URL,
			Data: base64.StdEncoding.EncodeToString(data),
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
			Code: 200,
			Url:  options.URL,
			Data: data,
		}.Bytes())
	})

	log.Println("listen at address:", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
