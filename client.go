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
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/retry"
)

const (
	retryFirstDelay   = 2 * time.Second
	retryMaxDelay     = 30 * time.Second
	retryFactor       = 2
	retryMaxAttempts  = 6
	itemsPerPage      = 500
	maxResponseLength = 1024 * 1024
)

var errResponseTooLarge = retry.PermanentError{
	Cause: errors.New("Response from CloudFlare is too large"),
}

func NewClient(creds Credentials, options ...Option) *Client {
	ret := Client{
		settings: applyOptions(options...),
		creds:    creds,
	}

	return &ret
}

type Client struct {
	*settings
	creds Credentials
}

// runWithRetry tries sending the request until it succeeds, fail to
// many times of fails once with a permanent error. Wait between retries
// use exponential backoff.
//
// This is not a method of Client because go allows using a type parameter
// on a method, but not declaring them.
func runWithRetry[TRESP commonResponseSetter](
	ctx context.Context,
	client *Client,
	logger *logs.Logger,
	req *request,
) (
	*response[TRESP],
	error,
) {
	var resp *response[TRESP]
	reterr := retry.ExpBackoff(ctx, logger, retryFirstDelay, retryMaxDelay,
		retryFactor, retryMaxAttempts, func() error {
			var err error
			resp, err = sendRequest[TRESP](ctx, client, logger, req)
			return err
		})

	return resp, reterr
}

// sendRequest sends an HTTP request, parses and returns the response.
// Permanent errors are wrapped with retry.PermanentError. Any error returned
// from the server is wrapped with HTTPError. If the error is a valid
// CloudFlare error, it is also wrapped with CloudFlareError.
func sendRequest[TRESP commonResponseSetter](
	ctx context.Context,
	client *Client,
	logger *logs.Logger,
	treq *request,
) (
	*response[TRESP],
	error,
) {
	err := client.ratelim.Wait(ctx)
	if err != nil {
		return nil, err
	}

	// request body
	var reqBody []byte
	if treq.body != nil {
		reqBody, err = json.Marshal(treq.body)
		if err != nil {
			return nil, retry.PermanentError{Cause: err}
		}
	}

	// request timeout
	reqCtx := ctx
	if client.requestTimeout > 0 {
		var reqCtxCancel func()
		reqCtx, reqCtxCancel = context.WithTimeout(reqCtx, client.requestTimeout)
		defer reqCtxCancel()
	}

	// create an http request
	req, err := http.NewRequestWithContext(reqCtx, treq.method, requestURL(treq),
		bytes.NewReader(reqBody))
	if err != nil {
		return nil, retry.PermanentError{Cause: err}
	}

	// headers
	req.Header.Set("Content-Type", "application/json")

	// credentials
	reqNoAuth := req.Clone(ctx)
	client.creds.configure(req)

	// send the request
	resp, err := client.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, retry.PermanentError{Cause: err}
		}

		// errors from Do() may be permanent or not, it is not possible to
		// determine precisely
		return nil, err // allow retry
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	// handle response
	if resp.StatusCode >= 400 {
		err = handleErrorResponse(resp, logger)
		logFullRequestError(logger, reqNoAuth, reqBody, err)
		return nil, err
	}

	tresp, err := handleSuccessResponse[TRESP](resp, logger)
	if err != nil {
		logFullRequestError(logger, reqNoAuth, reqBody, err)
		return nil, err
	}

	if client.logSuccess {
		logFullHTTPRequestSuccess(logger, reqNoAuth, reqBody, tresp)
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
	ret.rawBody, err = readResponseBody(httpResp.Body)
	if err != nil {
		// error response already specifies is can retry or not
		return nil, errors.Join(err, HTTPError{
			Code:    httpResp.StatusCode,
			RawBody: ret.rawBody,
			Headers: ret.headers,
		})
	}

	if len(ret.rawBody) == maxResponseLength {
		logger.W(fmt.Sprintf("Response from CloudFlare rejected because is bigger than %d", maxResponseLength))
		return nil, retry.PermanentError{
			Cause: errors.Join(err, HTTPError{
				Code:    httpResp.StatusCode,
				RawBody: ret.rawBody,
				Headers: ret.headers,
			}),
		}
	}

	err = json.Unmarshal(ret.rawBody, &ret.body)
	if err != nil {
		// error response already specifies is can retry or not
		return nil, errors.Join(err, HTTPError{
			Code:    httpResp.StatusCode,
			RawBody: ret.rawBody,
			Headers: ret.headers,
		})
	}

	return &ret, nil
}

func handleErrorResponse(resp *http.Response, _ *logs.Logger) error {
	// the error response must always support errors.As(err, HTTPError)
	httpErr := HTTPError{
		Code:    resp.StatusCode,
		Headers: resp.Header,
	}

	respBody, err := readResponseBody(resp.Body)
	if err != nil {
		err := fmt.Errorf("CloudFlare returned an error, but failed to read the error body: %w; %w", err, httpErr)

		if errors.Is(err, errResponseTooLarge) {
			return retry.PermanentError{Cause: err}
		}

		return err
	}

	httpErr.RawBody = respBody

	// try to parse the CloudFlare error objects
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if err != nil {
		return retry.PermanentError{Cause: fmt.Errorf("CloudFlare returned an error, and the content-type of the response is invalid %q (%w): %w",
			resp.Header.Get("content-type"), err, httpErr)}
	}

	if mediaType != "application/json" {
		return retry.PermanentError{Cause: fmt.Errorf("CloudFlare returned an error, and the content-type of the response is invalid %q: %w",
			resp.Header.Get("content-type"), httpErr)}
	}

	var cfcommon cfResponseCommon
	err = json.Unmarshal(respBody, &cfcommon)
	if err != nil {
		return retry.PermanentError{Cause: fmt.Errorf("CloudFlare returned an error, unmarshalling the error body as json failed: %w; %w", err, httpErr)}
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

func logFullRequestError(logger *logs.Logger, req *http.Request, reqBody []byte, err error) {
	logger.D(func(log logs.DebugFn) {
		msg := &strings.Builder{}

		// request

		fmt.Fprintln(msg, "REQUEST:")
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Fprintf(msg, "%s\n\n", reqDump)

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

func logFullHTTPRequestSuccess[TRESP commonResponseSetter](logger *logs.Logger, req *http.Request, reqBody []byte, resp *response[TRESP]) {
	logger.D(func(log logs.DebugFn) {
		msg := &strings.Builder{}

		// request
		fmt.Fprintln(msg, "REQUEST:")
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		reqDump, _ := httputil.DumpRequestOut(req, true)
		fmt.Fprintf(msg, "%s\n\n", reqDump)

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

func requestURL(treq *request) string {
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

func readResponseBody(body io.Reader) ([]byte, error) {
	ret, err := io.ReadAll(io.LimitReader(body, maxResponseLength+1))
	if err != nil {
		return nil, err // allow retry
	}

	if len(ret) > maxResponseLength {
		return nil, errResponseTooLarge // permanent error
	}

	return ret, nil
}
