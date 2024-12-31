package web

import (
	"encoding/json"
	"net/http"
)

func respondError(w http.ResponseWriter, code int, err string) {
	dto := map[string]string{
		"error": err,
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(dto)
}

func respondDTO(w http.ResponseWriter, dto interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(dto)
}
