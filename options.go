package cfdns

import (
	"net/http"
	"os"
	"time"

	"github.com/simplesurance/cfdns/logs"
	"github.com/simplesurance/cfdns/logs/textlogger"
	"golang.org/x/time/rate"
)

type Option func(*settings)

type settings struct {
	ratelim    *rate.Limiter
	logger     *logs.Logger
	httpClient *http.Client
}

func applyOptions(opts ...Option) *settings {
	ret := settings{
		ratelim:    rate.NewLimiter(rate.Every(defaultRequestInterval), 1),
		logger:     logs.FromDriver(textlogger.New(os.Stdout, os.Stderr), ""),
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(&ret)
	}

	return &ret
}

func WithRateLimit(interval time.Duration) Option {
	return func(s *settings) {
		s.ratelim = rate.NewLimiter(rate.Every(interval), 1)
	}
}

func WithLogger(logger *logs.Logger) Option {
	return func(s *settings) {
		s.logger = logger
	}
}
