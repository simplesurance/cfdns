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
	req request[TREQ],
) (
	resp response[TRESP],
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
	treq request[TREQ],
) (
	tresp response[TRESP],
	err error,
) {
	if err = treq.client.cfg.ratelim.Wait(ctx); err != nil {
		return
	}

	// url
	theurl, err := url.Parse(baseURL + "/" + treq.path)
	if err != nil {
		err = retry.PermanentError{Cause: err}
		return
	}

	theurl.RawQuery = treq.queryParams.Encode()

	// request body
	reqBody, err := json.Marshal(treq.body)
	if err != nil {
		err = retry.PermanentError{Cause: err}
		return
	}

	logger.D(func(log logs.DebugFn) {
		log(fmt.Sprintf("Request body: %s", reqBody))
	})

	// request
	req, err := http.NewRequestWithContext(ctx, treq.method, theurl.String(),
		bytes.NewReader(reqBody))
	if err != nil {
		err = retry.PermanentError{Cause: err}
		return
	}

	// headers
	mergeHeaders(req.Header, treq.headers)

	// credentials
	err = treq.client.creds.configure(req)
	if err != nil {
		return // allow retry
	}

	// send the request
	logger.D(func(log logs.DebugFn) {
		log(treq.method + " " + theurl.String())
	})
	resp, err := treq.client.cfg.httpClient.Do(req)
	if err != nil {
		// errors from Do() may be permanent or not, it is not possible to
		// determine precisely
		return // allow retry
	}

	// handle response
	if resp.StatusCode >= 400 {
		err = handleErrorResponse(resp, logger)

		var httpErr HTTPError
		if errors.As(err, &httpErr) && httpErr.IsPermanent() {
			err = retry.PermanentError{Cause: err}
			return
		}

		return
	}

	return handleSuccessResponse[TRESP](resp, logger)
}

func handleSuccessResponse[TRESP commonResponseSetter](httpResp *http.Response, logger *logs.Logger) (
	resp response[TRESP],
	err error,
) {
	resp.code = httpResp.StatusCode
	resp.headers = httpResp.Header

	var respBody []byte
	respBody, err = io.ReadAll(httpResp.Body)
	if err != nil {
		err = errors.Join(err, HTTPError{
			Code:    httpResp.StatusCode,
			RawBody: respBody,
			Headers: resp.headers,
		})
		return
	}

	err = json.Unmarshal(respBody, &resp.body)
	if err != nil {
		err = errors.Join(err, HTTPError{
			Code:    httpResp.StatusCode,
			RawBody: respBody,
			Headers: resp.headers,
		})
		return
	}

	return
}

func handleErrorResponse(resp *http.Response, logger *logs.Logger) error {
	// the error response must always support errors.As(err, HTTPError)
	httpErr := HTTPError{
		Code:    resp.StatusCode,
		Headers: resp.Header,
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("CloudFlare returned an error, but failed to read the error body: %v; %w", err, httpErr)
	}

	httpErr.RawBody = respBody

	// try to parse the CloudFlare error objects
	mediaType, _, err := mime.ParseMediaType(resp.Header.Get("content-type"))
	if err != nil {
		return fmt.Errorf("CloudFlare returned an error, and the content-type of the response is unsupported %q (%v); %w",
			resp.Header.Get("content-type"), err, httpErr)
	}

	if mediaType != "application/json" {
		return fmt.Errorf("CloudFlare returned an error, and the content-type of the response is unsupported %q\n%w",
			resp.Header.Get("content-type"), httpErr)
	}

	var cfcommon cfResponseCommon
	err = json.Unmarshal(respBody, &cfcommon)
	if err != nil {
		return fmt.Errorf("CloudFlare returned an error, but failed to read the error body: %v; %w", err, httpErr)
	}

	return CloudFlareError{
		cfResponseCommon: cfcommon,
		HTTPError:        httpErr,
	}
}

// mergeHeaders add the values on the second parameter to the first.
func mergeHeaders(dst, target http.Header) {
	for k, v := range target {
		dst[k] = v
	}
}
