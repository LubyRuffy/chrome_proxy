package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestResult_Bytes(t *testing.T) {
	debug = true
	data, err := renderURLDOM(&chromeParam{
		URL:     "https://bgp.he.net/ip/106.75.29.24",
		Sleep:   5,
		Timeout: 30,
	})
	assert.Nil(t, err)
	assert.Contains(t, data, "fofa.info")
}
