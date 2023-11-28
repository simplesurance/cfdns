package cfdns

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/exp/maps"
)

type request struct {
	method      string
	path        string
	queryParams url.Values
	body        any // the encoding/json package will be used to marshal it
}

type response[T any] struct {
	body    T
	rawBody []byte
	code    int
	headers http.Header
}

type HTTPError struct {
	Code    int
	RawBody []byte
	Headers http.Header
}

func (e HTTPError) Error() string {
	msg := &strings.Builder{}

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
	errs := make([]string, len(ce.Errors))
	for i, err := range ce.Errors {
		var chain []string
		for _, ch := range err.ErrorChain {
			chain = append(chain, fmt.Sprintf("%d %s", ch.Code, ch.Message))
		}

		chainmsg := ""
		if len(chain) > 0 {
			chainmsg = fmt.Sprintf(" (%s)", strings.Join(chain, "; "))
		}

		errs[i] = fmt.Sprintf("%d %s%s", err.Code, err.Message, chainmsg)
	}

	return fmt.Sprintf("CloudFlare error: %s\n\n%s", strings.Join(errs, ", "), ce.HTTPError.Error())
}

func (ce CloudFlareError) Unwrap() error {
	return ce.HTTPError
}

var _ error = CloudFlareError{}
