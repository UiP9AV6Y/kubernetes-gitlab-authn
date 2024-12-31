package web

import (
	"net/http"
	"strings"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderPrivateToken  = "PRIVATE-TOKEN"
)

func parseToken(r *http.Request) string {
	token := r.Header.Get(HeaderPrivateToken)
	if token != "" {
		return token
	}

	auth := r.Header.Get(HeaderAuthorization)
	if auth == "" {
		return auth
	}

	fields := strings.Fields(auth)
	if len(fields) != 2 {
		return ""
	}

	// we ignore the auth type for simplicity sake
	return fields[1]
}
