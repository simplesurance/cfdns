package cfdns

import "net/http"

type Credentials interface {
	configure(*http.Request)
}

func APIToken(token string) Credentials {
	return apiToken{token: token}
}

type apiToken struct {
	token string
}

func (a apiToken) configure(req *http.Request) {
	panic("TODO implement")
}

var _ Credentials = apiToken{}
