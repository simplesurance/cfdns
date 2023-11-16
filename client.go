package cfdns

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
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

			if err != nil {
				panic(err)
			}
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
	logger.D("Sending HTTP request to " + theurl.String())
	resp, err := treq.client.cfg.httpClient.Do(req)
	if err != nil {
		// errors from Do() may be permanent or not, it is not possible to
		// determine precisely; allowing retry
		return
	}

	// handle response
	if resp.StatusCode >= 400 {
		err = handleErrorResponse(resp)
		return
	}

	return handleSuccessResponse[TRESP](resp)
}

func handleSuccessResponse[TRESP commonResponseSetter](httpResp *http.Response) (
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

func handleErrorResponse(resp *http.Response) error {
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	httpErr := HTTPError{
		Code:    resp.StatusCode,
		RawBody: respBody,
		Headers: resp.Header,
	}

	return httpErr
}

// mergeHeaders add the values on the second parameter to the first.
func mergeHeaders(dst, target http.Header) {
	for k, v := range target {
		dst[k] = v
	}
}
