package cfdns

import (
	"fmt"
	"net/http"
	"net/url"
)

type request[T any] struct {
	client      *Client
	method      string
	path        string
	queryParams url.Values
	headers     http.Header
	body        T
}

type response[T any] struct {
	body    T
	code    int
	headers http.Header
}

type HTTPError struct {
	Code    int
	RawBody []byte
	Headers http.Header
}

func (e HTTPError) Error() string {
	// TODO add more details
	return fmt.Sprintf("HTTP %d\n%s", e.Code, e.RawBody)
}

var _ error = HTTPError{}
