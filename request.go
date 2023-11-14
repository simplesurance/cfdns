package cfdns

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/simplesurance/cfdns/logs"
)

type request[T any] struct {
	logger      logs.Logger
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
	headers url.Values
}

type HTTPError struct {
	Code    int
	RawBody []byte
	Headers http.Header
}

func (e HTTPError) Error() string {
	// TODO add more details
	return fmt.Sprintf("server response is %d")
}

var _ error = HTTPError{}
