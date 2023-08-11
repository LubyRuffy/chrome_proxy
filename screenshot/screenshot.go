package screenshot

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/LubyRuffy/chrome_proxy/chrome_action"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/LubyRuffy/chrome_proxy/utils"
	"github.com/chromedp/chromedp"
	"log"
	"os"
	"time"
)

// ScreenshotURL 截图
func ScreenshotURL(options *models.ChromeParam) (*models.ScreenshotOutput, error) {
	log.Println("screenshot of url:", options.URL)

	var buf []byte
	var actions []chromedp.Action
	actions = append(actions, chromedp.CaptureScreenshot(&buf))

	var title string
	actions = append(actions, chromedp.Title(&title))
	var url string
	actions = append(actions, chromedp.Location(&url))

	err := chrome_action.ChromeActions(options.ChromeActionInput, func(s string, i ...interface{}) {

	}, options.Timeout, nil, actions...)
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

	return &models.ScreenshotOutput{
		Data:     buf,
		Title:    title,
		Location: url,
	}, err
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
	fn, err = utils.WriteTempFile(".html", func(f *os.File) error {
		_, err = f.Write(html)
		return err
	})

	// 将html文件进行截图
	ctx, cancel := chromedp.NewContext(
		context.Background(),
	)
	defer cancel()

	var buf []byte
	if err = chromedp.Run(ctx, FullScreenshot(`file://`+fn, 100, &buf)); err != nil {
		return nil, err
	}

	return buf, err
}

// FullScreenshot takes a screenshot of the entire browser viewport.
//
// Note: chromedp.FullScreenshot overrides the device's emulation settings. Use
// device.Reset to reset the emulation and viewport settings.
func FullScreenshot(urlstr string, quality int, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.CaptureScreenshot(res),
	}
}
