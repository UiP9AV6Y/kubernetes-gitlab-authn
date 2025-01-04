package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderPrivateToken  = "PRIVATE-TOKEN"

	QueryParamPagination = "pagination"
	QueryParamOrder      = "order_by"
	QueryParamSize       = "per_page"
	QueryParamPage       = "page"
	QueryParamSort       = "sort"

	DefaultPage      = 1
	DefaultBatchSize = 20
	MaxBatchSize     = 100
)

func parseRequestURL(r *http.Request) *url.URL {
	u := *(r.URL)
	u.RawQuery = ""
	u.Fragment = ""
	u.RawFragment = ""

	return &u
}

func parseOffsetPagination(r *http.Request) (page int, size int) {
	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return DefaultPage, DefaultBatchSize
	}

	p := q.Get(QueryParamPage)
	if p != "" {
		page, _ = strconv.Atoi(p)
	}
	if page <= 0 {
		page = DefaultPage
	}

	s := q.Get(QueryParamSize)
	if s != "" {
		size, _ = strconv.Atoi(s)
	}
	if size <= 0 || size > MaxBatchSize {
		size = DefaultBatchSize
	}

	return
}

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
