package cfdns

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/retry"
)

func NewClient(creds Credentials, options ...Option) *Client {
	ret := Client{
		cfg: applyOptions(options...),
	}

	return &ret
}

type Client struct {
	cfg *settings
}

func run[T any](
	ctx context.Context,
	logger logs.Logger,
	client *Client,
	method string,
	path string,
	queryParams url.Values,
	headers http.Header,
	body any,
	withRetry bool,
) (
	respBody T,
	code int,
	respHeaders url.Values,
	_ error,
) {
	err := retry.ExpBackoff(ctx, logger, 1, 30, 1.5, 5, func() error {
		return nil
	})
}

func runOnce[T any](
	ctx context.Context,
	client Client,
	method string,
	path string,
	queryParams url.Values,
	headers http.Header,
	body any,
) (
	respBody T,
	code int,
	respHeaders url.Values,
	_ error,
) {
	var nilT T

	// url
	theurl, err := url.Parse(baseURL + path)
	if err != nil {
		return nilT, 0, nil, retry.PermanentError{Cause: err}
	}

	theurl.RawQuery = queryParams.Encode()

	// request body
	reqBody, err := json.Marshal(body)
	if err != nil {
		return nilT, 0, nil, retry.PermanentError{Cause: err}
	}

	// request
	req, err := http.NewRequestWithContext(ctx, method, theurl.String(),
		bytes.NewReader(reqBody))
	if err != nil {
		return nilT, 0, nil, retry.PermanentError{Cause: err}
	}

	// headers
	mergeHeaders(req.Header, headers)

	// send the request
	resp, err := client.cfg.httpClient.Do(req)
	if err != nil {
		return nilT, 0, nil, retry.PermanentError{Cause: err}
	}

}

// mergeHeaders add the values on the second parameter to the first.
func mergeHeaders(dst, target http.Header) {
	for k, v := range target {
		dst[k] = v
	}
}
