package cfdns

import (
	"net/http"
	"os"

	"golang.org/x/time/rate"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/textlogger"
)

type Option func(*settings)

type settings struct {
	ratelim    *rate.Limiter
	logger     *logs.Logger
	httpClient *http.Client
	logSuccess bool
}

func applyOptions(opts ...Option) *settings {
	ret := settings{
		ratelim:    rate.NewLimiter(rate.Every(defaultRequestInterval), 1),
		logger:     logs.New(textlogger.New(os.Stdout, os.Stderr)),
		httpClient: http.DefaultClient,
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

func WithLogger(logger *logs.Logger) Option {
	return func(s *settings) {
		s.logger = logger
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
