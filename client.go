package cfdns

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

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

func runWithRetry[TREQ, TRESP any](
	ctx context.Context,
	req request[TREQ],
) (
	resp response[TRESP],
	_ error,
) {
	reterr := retry.ExpBackoff(ctx, req.logger, 1, 30, 1.5, 5, func() error {
		var err error
		resp, err = runOnce[TREQ, TRESP](ctx, req)

		return err
	})

	return resp, reterr
}

// runOnce sends an HTTP request, parses and returns the response.
// Permanent errors are wrapped with retry.PermanentError. Any error returned
// from the server is wrapped with HTTPError. If the error is a valid
// CloudFlare error, it is also wrapped with CloudFlareError.
func runOnce[TREQ, TRESP any](
	ctx context.Context,
	treq request[TREQ],
) (
	tresp response[TRESP],
	err error,
) {
	// url
	theurl, err := url.Parse(baseURL + treq.path)
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

	// send the request
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

func handleErrorResponse(resp *http.Response) (err error) {
}

func handleSuccessResponse[TRESP any](resp *http.Response) (
	tresp response[TRESP],
	err error,
) {
}

// mergeHeaders add the values on the second parameter to the first.
func mergeHeaders(dst, target http.Header) {
	for k, v := range target {
		dst[k] = v
	}
}
