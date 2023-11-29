package cfdns

import (
	"errors"
	"net/http"
)

// ErrEmptyToken is returned when the credentials generator produces an empty
// authentication token.
var ErrEmptyToken = errors.New("Provided token is empty")

type Credentials interface {
	configure(*http.Request)
}

func APIToken(token string) (Credentials, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	return apiToken{token: token}, nil
}

type apiToken struct {
	token string
}

func (a apiToken) configure(req *http.Request) {
	req.Header.Set("authorization", "Bearer "+a.token)
}

var _ Credentials = apiToken{}
