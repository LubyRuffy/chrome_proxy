package utils

import (
	"encoding/json"
	"github.com/LubyRuffy/chrome_proxy/models"
	"net/http"
)

func GetOptionFromRequest(r *http.Request) (*models.ChromeParam, error) {
	var options models.ChromeParam
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
