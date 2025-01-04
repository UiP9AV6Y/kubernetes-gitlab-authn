package web

import (
	"encoding/json"
	"net/http"
)

const (
	HeaderContentType  = "Content-Type"
	HeaderNextPage     = "X-Next-Page"
	HeaderCurrentPage  = "X-Page"
	HeaderPreviousPage = "X-Prev-Page"
	HeaderTotalPages   = "X-Total-Pages"
	HeaderPageSize     = "X-Per-Page"
	HeaderTotalItems   = "X-Total"
	HeaderLink         = "Link"
)

const (
	ContentTypeJSON = "application/json; charset=utf-8"
)

func respondError(w http.ResponseWriter, code int, err string) {
	dto := map[string]string{
		"error": err,
	}
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(dto)
}

func respondDTO(w http.ResponseWriter, dto interface{}) {
	w.Header().Set(HeaderContentType, ContentTypeJSON)
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto)
}
