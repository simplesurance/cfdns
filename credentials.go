package cfdns

import (
	"errors"
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

func (a apiToken) configure(_ *logs.Logger, req *http.Request) error {
	if a.token == "" {
		return errEmptyToken
	}

	req.Header.Set("authorization", "Bearer "+a.token)

	return nil
}

var _ Credentials = apiToken{}
