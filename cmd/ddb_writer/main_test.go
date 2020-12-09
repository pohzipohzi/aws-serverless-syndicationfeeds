package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/mmcdole/gofeed"
	"github.com/stretchr/testify/assert"
)

func Test_handler(t *testing.T) {
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/xml")
		io.WriteString(w, `<rss version="2.0"></rss>`)
	}))
	fp := gofeed.NewParser()
	fp.Client = &http.Client{Transport: &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			return url.Parse(testServer.URL)
		},
	}}
	h := &handler{
		feedParser: fp,
	}
	assert.Nil(t, h.Handle(input{URL: testServer.URL}))
}
