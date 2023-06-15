package main

import (
	"bytes"
	"embed"
	_ "embed"
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/colornames"
	"image"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"
)

//go:embed resource/font.ttf
//go:embed resource/sample.png
var f embed.FS

func addNavBar(url, title string, picBuf []byte) ([]byte, error) {
	// 解码截图图像
	img, _, err := image.Decode(bytes.NewReader(picBuf))
	if err != nil {
		log.Fatal(err)
	}

	// 生成导航页
	if len(title) > 36 {
		title = title[:35]
	}
	if len(url) > 176 {
		url = url[:175]
	}

	navImg, err := drawNavInfo(title, url)
	if err != nil {
		return nil, err
	}

	// 打开导航页图
	sampleImg, err := imaging.Open(navImg)
	if err != nil {
		log.Fatal(err)
	}
	// 调整导航页图的大小以适配截图的宽度
	sampleImg = imaging.Resize(sampleImg, img.Bounds().Dx(), 0, imaging.Lanczos)

	// 创建一个新的图像，大小为截图宽度 x (截图高度 + 导航页图高度)
	canvas := image.NewRGBA(image.Rect(0, 0, img.Bounds().Dx(), img.Bounds().Dy()+sampleImg.Bounds().Dy()))

	// 将导航页图绘制在画布上方
	draw.Draw(canvas, sampleImg.Bounds(), sampleImg, image.ZP, draw.Over)

	// 将截图绘制在画布下方
	draw.Draw(canvas, img.Bounds().Add(image.Pt(0, sampleImg.Bounds().Dy())), img, image.ZP, draw.Over)

	// 创建一个字节缓冲区
	buf := new(bytes.Buffer)
	err = png.Encode(buf, canvas)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// 根据文件扩展名将图像保存为 JPEG 或 PNG
func saveImageWithFormat(file *os.File, img image.Image, path string) error {
	ext := filepath.Ext(path)
	switch strings.ToLower(ext) {
	case ".jpeg", ".jpg":
		return jpeg.Encode(file, img, &jpeg.Options{Quality: 100})
	case ".png":
		return png.Encode(file, img)
	default:
		return fmt.Errorf("unsupported file format: %s", ext)
	}
}

const (
	SampleImagePath = "resource/sample.png"
	FontSize        = 7
	URLFontSize     = 8
	FirstLineX      = 51
	FirstLineY      = 13
	SecondLineX     = 58
	SecondLineY     = 27.5
)

func drawNavInfo(title, url string) (string, error) {
	// 打开 sample.jpg 图片文件
	file, err := f.ReadFile(SampleImagePath)
	if err != nil {
		log.Fatal(err)
	}

	// 解码图像
	img, _, err := image.Decode(bytes.NewReader(file))
	if err != nil {
		log.Fatal(err)
	}

	// 创建一个新的图像，大小与原图相同
	dst := image.NewRGBA(img.Bounds())
	draw.Draw(dst, img.Bounds(), img, image.Point{}, draw.Src)

	// 创建绘图上下文
	dc := gg.NewContextForRGBA(dst)

	// 加载字体文件
	fontData, err := f.ReadFile("resource/font.ttf")
	if err != nil {
		log.Fatal(err)
	}

	// 解析字体文件
	font, err := truetype.Parse(fontData)
	if err != nil {
		log.Fatal(err)
	}

	// 创建字体
	fontFace := truetype.NewFace(font, &truetype.Options{
		Size: FontSize,
	})

	// 设置字体
	dc.SetFontFace(fontFace)

	// 设置文本颜色为黑色
	dc.SetColor(colornames.Black)

	// 绘制图片
	dc.DrawImage(img, 0, 0)

	// 绘制第一行文本
	dc.DrawString(title, FirstLineX, FirstLineY)

	// 创建字体
	fontFace = truetype.NewFace(font, &truetype.Options{
		Size: URLFontSize,
	})

	// 设置字体
	dc.SetFontFace(fontFace)

	// 设置文本颜色为黑色
	dc.SetColor(colornames.Black)

	// 绘制第二行文本
	dc.DrawString(url, SecondLineX, SecondLineY)

	temp, err := os.CreateTemp("", "chrome_proxy_*.jpg")
	if err != nil {
		return "", err
	}

	outfile := temp
	err = saveImageWithFormat(outfile, dc.Image(), temp.Name())
	if err != nil {
		return "", err
	}

	return outfile.Name(), nil
}
