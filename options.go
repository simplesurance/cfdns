package cfdns

import (
	"net/http"
	"time"

	"golang.org/x/time/rate"

	"github.com/simplesurance/cfdns/log"
	"github.com/simplesurance/cfdns/log/niltarget"
)

type Option func(*settings)

type settings struct {
	ratelim        *rate.Limiter
	logger         *log.Logger
	httpClient     *http.Client
	logSuccess     bool
	requestTimeout time.Duration
}

func applyOptions(opts ...Option) *settings {
	ret := settings{
		ratelim:        rate.NewLimiter(rate.Every(defaultRequestInterval), 1),
		logger:         log.New(niltarget.New()), // by default log messages are suppressed
		httpClient:     http.DefaultClient,
		requestTimeout: 30 * time.Second,
	}
	for _, opt := range opts {
		opt(&ret)
	}

	return &ret
}

func WithRateLimiter(ratelim *rate.Limiter) Option {
	return func(s *settings) {
		s.ratelim = ratelim
	}
}

func WithLogger(logger *log.Logger) Option {
	return func(s *settings) {
		s.logger = logger
	}
}

func WithHTTPClient(c *http.Client) Option {
	return func(s *settings) {
		s.httpClient = c
	}
}

// WithRequestTimeout configures how long to wait for an HTTP request.
// The default is 1 minute. Setting a value of 0 will make it use the
// default behavior of the HTTP client being used, that might be
// waiting forever.
func WithRequestTimeout(timeout time.Duration) Option {
	return func(s *settings) {
		s.requestTimeout = timeout
	}
}

// WithLogSuccessfulResponses allows logging full request and response
// send to CloudFlare in case of successful response. Debug log must also
// be enabled.
//
// Error responses will always be logged if debug log is enabled.
func WithLogSuccessfulResponses(enable bool) Option {
	return func(s *settings) {
		s.logSuccess = enable
	}
}
