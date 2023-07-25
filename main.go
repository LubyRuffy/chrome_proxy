package main

import (
	"encoding/base64"
	"flag"
	"github.com/LubyRuffy/chrome_proxy/models"
	"github.com/LubyRuffy/chrome_proxy/render_dom"
	"github.com/LubyRuffy/chrome_proxy/screenshot"
	"github.com/LubyRuffy/chrome_proxy/utils"
	"log"
	"net/http"
)

func main() {
	addr := flag.String("addr", ":5558", "http server listen address")
	flag.Parse()

	http.HandleFunc("/screenshot", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		options, err := utils.GetOptionFromRequest(r)
		if options == nil {
			w.Write(models.Result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		screenshotResult, err := screenshot.ScreenshotURL(options)
		if err != nil {
			w.Write(models.Result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(models.Result{
			Code:     200,
			Url:      options.URL,
			Data:     base64.StdEncoding.EncodeToString(screenshotResult.Data),
			Title:    screenshotResult.Title,
			Location: screenshotResult.Location,
		}.Bytes())
	})

	http.HandleFunc("/renderDom", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		options, err := utils.GetOptionFromRequest(r)
		if options == nil {
			w.Write(models.Result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}

		data, err := render_dom.RenderDom(options)
		if err != nil {
			w.Write(models.Result{
				Code:    500,
				Message: err.Error(),
			}.Bytes())
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(models.Result{
			Code:     200,
			Url:      options.URL,
			Data:     data.Html,
			Title:    data.Title,
			Location: data.Location,
		}.Bytes())
	})

	log.Println("listen at address:", *addr)
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
