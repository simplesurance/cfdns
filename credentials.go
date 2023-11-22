package cfdns

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/simplesurance/cfdns/logs"
)

var errEmptyToken = errors.New("Provided token is empty")

type Credentials interface {
	configure(*logs.Logger, *http.Request) error
}

func APIToken(token string) Credentials {
	return apiToken{token: token}
}

type apiToken struct {
	token string
}

func (a apiToken) configure(logger *logs.Logger, req *http.Request) error {
	if a.token == "" {
		return errEmptyToken
	}

	req.Header.Set("authorization", "Bearer "+a.token)

	// FIXME remove
	logger.D(func(log logs.DebugFn) {
		log(fmt.Sprintf("APIToken injected"))
	})

	return nil
}

var _ Credentials = apiToken{}
