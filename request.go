package cfdns

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/exp/maps"
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
	msg := &bytes.Buffer{}

	fmt.Fprintf(msg, "HTTP %d\n", e.Code)
	headers := maps.Keys(e.Headers)
	for _, k := range headers {
		fmt.Fprintf(msg, "%s: %s\n", k, e.Headers.Get(k))
	}
	fmt.Fprintln(msg)
	fmt.Fprintf(msg, "%s", e.RawBody)

	return msg.String()
}

// IsPermanent returns true if should not try again the same request.
func (e HTTPError) IsPermanent() bool {
	if e.Code >= 400 || e.Code < 500 {
		return true
	}

	return false
}

var _ error = HTTPError{}

type CloudFlareError struct {
	cfResponseCommon
	HTTPError HTTPError
}

func (ce CloudFlareError) Error() string {
	var errs []string
	for _, err := range ce.Errors {
		var chain []string
		for _, ch := range err.ErrorChain {
			chain = append(chain, fmt.Sprintf("%d %s", ch.Code, ch.Message))
		}

		chainmsg := ""
		if len(chain) > 0 {
			chainmsg = fmt.Sprintf(" (%s)", strings.Join(chain, "; "))
		}

		errs = append(errs, fmt.Sprintf("%d %s%s", err.Code, err.Message, chainmsg))
	}

	return fmt.Sprintf("CloudFlare error: %s\n\n%s", strings.Join(errs, ", "), ce.HTTPError.Error())
}

func (ce CloudFlareError) Unwrap() error {
	return ce.HTTPError
}

var _ error = CloudFlareError{}