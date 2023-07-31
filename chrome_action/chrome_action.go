package chrome_action

import (
	"context"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"strings"
	"time"
)

// ChromeActions 完成chrome的headless操作
func ChromeActions(in models.ChromeActionInput, logf func(string, ...interface{}), timeout int, actions ...chromedp.Action) error {
	var err error

	// set user-agent
	if in.UserAgent == "" {
		in.UserAgent = models.DefaultUserAgent
	}

	// prepare the chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
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

	if models.Debug {
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
