package cfdns

import "net/http"

type Credentials interface {
	configure(*http.Request) error
}

func APIToken(token string) Credentials {
	return apiToken{token: token}
}

type apiToken struct {
	token string
}

func (a apiToken) configure(req *http.Request) error {
	req.Header.Set("authorization", "Bearer "+a.token)
	return nil
}

var _ Credentials = apiToken{}
