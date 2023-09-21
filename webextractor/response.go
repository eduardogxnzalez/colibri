package webextractor

import (
	"io"
	"net/http"
	"net/url"

	"github.com/eduardogxnzalez/colibri"
)

// Response represents an HTTP response.
// See the colibri.Response interface.
type Response struct {
	HTTP *http.Response
	c    *colibri.Colibri
}

func (resp *Response) URL() *url.URL {
	return resp.HTTP.Request.URL
}

func (resp *Response) StatusCode() int {
	return resp.HTTP.StatusCode
}

func (resp *Response) Header() http.Header {
	return resp.HTTP.Header
}

func (resp *Response) Body() io.ReadCloser {
	return resp.HTTP.Body
}

func (resp *Response) Do(rules *colibri.Rules) (colibri.Response, error) {
	return resp.c.Do(rules)
}

func (resp *Response) Extract(rules *colibri.Rules) (colibri.Response, map[string]any, error) {
	return resp.c.Extract(rules)
}
