package cfdns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/retry"
)

const (
	retryFirsDelay   = 2 * time.Second
	retryMaxDelay    = 30 * time.Second
	retryFactor      = 2
	retryMaxAttempts = 6
	itemsPerPage     = 500
)

func NewClient(creds Credentials, options ...Option) *Client {
	ret := Client{
		cfg:   applyOptions(options...),
		creds: creds,
	}

	return &ret
}

type Client struct {
	cfg   *settings
	creds Credentials
}

func runWithRetry[TREQ any, TRESP commonResponseSetter](
	ctx context.Context,
	logger *logs.Logger,
	req *request[TREQ],
) (
	resp *response[TRESP],
	_ error,
) {
	reterr := retry.ExpBackoff(ctx, logger, retryFirsDelay, retryMaxDelay,
		retryFactor, retryMaxAttempts, func() error {
			var err error
			resp, err = runOnce[TREQ, TRESP](ctx, logger, req)
			return err
		})

	return resp, reterr
}

// runOnce sends an HTTP request, parses and returns the response.
// Permanent errors are wrapped with retry.PermanentError. Any error returned
// from the server is wrapped with HTTPError. If the error is a valid
// CloudFlare error, it is also wrapped with CloudFlareError.
func runOnce[TREQ any, TRESP commonResponseSetter](
	ctx context.Context,
	logger *logs.Logger,
	treq *request[TREQ],
) (
	tresp *response[TRESP],
	err error,
) {
	if err = treq.client.cfg.ratelim.Wait(ctx); err != nil {
		return
	}

	// request body
	var reqBody []byte
	if treq.body != nil {
		reqBody, err = json.Marshal(treq.body)
		if err != nil {
			err = retry.PermanentError{Cause: err}
			return
		}
	}

	// request
	req, err := http.NewRequestWithContext(ctx, treq.method, requestURL(treq),
		bytes.NewReader(reqBody))
	if err != nil {
		err = retry.PermanentError{Cause: err}
		return
	}

	// headers
	mergeHeaders(req.Header, treq.headers)

	// credentials
	err = treq.client.creds.configure(
		logger.SubLogger(logs.WithPrefix("Authorization")),
		req)
	if err != nil {
		return // allow retry
	}

	// send the request
	resp, err := treq.client.cfg.httpClient.Do(req)
	if err != nil {
		// errors from Do() may be permanent or not, it is not possible to
		// determine precisely
		return // allow retry
	}

	// handle response
	if resp.StatusCode >= 400 {
		err = handleErrorResponse(resp, logger)
		logFullRequestError(logger, treq, reqBody, err)
		return
	}

	tresp, err = handleSuccessResponse[TRESP](resp, logger)
	if err != nil {
		logFullRequestError(logger, treq, reqBody, err)
		return
	}

	if treq.client.cfg.logSuccess {
		logFullHTTPRequestSuccess(logger, treq, reqBody, tresp)
	}

	return tresp, err
}

func handleSuccessResponse[TRESP commonResponseSetter](httpResp *http.Response, logger *logs.Logger) (
	*response[TRESP],
	error,
) {
	var ret response[TRESP]

	ret.code = httpResp.StatusCode
	ret.headers = httpResp.Header

	var err error
	ret.rawBody, err = io.ReadAll(httpResp.Body)
	if err != nil {
		// allow retry
		return nil, errors.Join(err, HTTPError{
			Code:    httpResp.StatusCode,
			RawBody: ret.rawBody,
			Headers: ret.headers,
		})
	}

	err = json.Unmarshal(ret.rawBody, &ret.body)
	if err != nil {
		// allow retry
		return nil, errors.Join(err, HTTPError{
			Code:    httpResp.StatusCode,
			RawBody: ret.rawBody,
			Headers: ret.headers,
		})
	}

	return &ret, nil
}

func handleErrorResponse(resp *http.Response, logger *logs.Logger) error {
	// the error response must always support errors.As(err, HTTPError)
	httpErr := HTTPError{
		Code:    resp.StatusCode,
		Headers: resp.Header,
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("CloudFlare returned an error, but failed to read the error body: %v; %w", err, httpErr) // allow retry
	}

	httpErr.RawBody = respBody

	// try to parse the CloudFlare error objects
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if err != nil {
		return retry.PermanentError{Cause: fmt.Errorf("CloudFlare returned an error, and the content-type of the response is unsupported %q (%v); %w",
			resp.Header.Get("content-type"), err, httpErr)}
	}

	if mediaType != "application/json" {
		return retry.PermanentError{Cause: fmt.Errorf("CloudFlare returned an error, and the content-type of the response is unsupported %q\n%w",
			resp.Header.Get("content-type"), httpErr)}
	}

	var cfcommon cfResponseCommon
	err = json.Unmarshal(respBody, &cfcommon)
	if err != nil {
		return retry.PermanentError{Cause: fmt.Errorf("CloudFlare returned an error, but failed to read the error body: %v; %w", err, httpErr)}
	}

	ret := CloudFlareError{
		cfResponseCommon: cfcommon,
		HTTPError:        httpErr,
	}

	if httpErr.IsPermanent() {
		return retry.PermanentError{Cause: ret}
	}

	return ret
}

func logFullRequestError[T any](logger *logs.Logger, treq *request[T], reqBody []byte, err error) {
	logger.D(func(log logs.DebugFn) {
		msg := &bytes.Buffer{}

		// request
		fmt.Fprintln(msg, "REQUEST:")
		fmt.Fprintf(msg, "%s %s\n", treq.method, requestURL(treq))
		for k, v := range treq.headers {
			fmt.Fprintf(msg, "%s: %s\n", k, strings.Join(v, ", "))
		}
		fmt.Fprintln(msg)
		if treq.body != nil {
			fmt.Fprintf(msg, "%s\n\n", reqBody)
		}

		var httpErr HTTPError
		if errors.As(err, &httpErr) {
			fmt.Fprintf(msg, "RESPONSE: %d\n", httpErr.Code)
			for k, v := range httpErr.Headers {
				fmt.Fprintf(msg, "%s: %s\n", k, strings.Join(v, ", "))
			}
			fmt.Fprintln(msg)
			fmt.Fprintf(msg, "%s", httpErr.RawBody)
		} else {
			// for the cases where we didn't go a response from the server
			fmt.Fprintf(msg, "Response: Go error %T: %v", err, err)
		}

		log("Failed REST call to CloudFlare:\n" + msg.String())
	})
}

func logFullHTTPRequestSuccess[TREQ any, TRESP commonResponseSetter](logger *logs.Logger, treq *request[TREQ], reqBody []byte, resp *response[TRESP]) {
	logger.D(func(log logs.DebugFn) {
		msg := &bytes.Buffer{}

		// request
		fmt.Fprintln(msg, "REQUEST:")
		fmt.Fprintf(msg, "%s %s\n", treq.method, requestURL(treq))
		for k, v := range treq.headers {
			fmt.Fprintf(msg, "%s: %s\n", k, strings.Join(v, ", "))
		}
		fmt.Fprintln(msg)
		if treq.body != nil {
			fmt.Fprintf(msg, "%s\n\n", reqBody)
		}

		// response
		fmt.Fprintf(msg, "RESPONSE: %d\n", resp.code)
		for k, v := range resp.headers {
			fmt.Fprintf(msg, "%s: %s\n", k, strings.Join(v, ", "))
		}
		fmt.Fprintln(msg)
		fmt.Fprintf(msg, "%s", resp.rawBody)

		log("Successful request to CloudFlare:\n" + msg.String())
	})
}

// mergeHeaders add the values on the second parameter to the first. In case
// of duplications, the second parameter "wins".
func mergeHeaders(dst, target http.Header) {
	for k, v := range target {
		dst[k] = v
	}
}

func requestURL[T any](treq *request[T]) string {
	urlstring := baseURL + "/" + treq.path
	theurl, err := url.Parse(urlstring)
	if err != nil {
		// this only happens in case of coding error on cfapi
		panic(fmt.Sprintf("URL %q is invalid; created from request %v",
			urlstring, treq))
	}

	theurl.RawQuery = treq.queryParams.Encode()

	return theurl.String()
}
