package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

const (
	HeaderAuthorization = "Authorization"
	HeaderPrivateToken  = "PRIVATE-TOKEN"
)

func notFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		slog.Warn("Missing route implementation", "path", req.URL.Path)
		respondError(w, http.StatusNotFound, "not implemented")
	}
}

func meHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			slog.Info("User request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		pk := findPK(auth)
		result, ok := userDAO[pk]
		if ok {
			slog.Info("User request yielded result", "principal", pk)
			respondDTO(w, result)
			return
		}

		slog.Info("User request can not be served properly", "token", auth)
		respondError(w, http.StatusNotFound, "user not found")
	}
}

func groupsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			slog.Info("Groups request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		pk := findPK(auth)
		result, ok := groupDAO[pk]
		if ok {
			slog.Info("Groups request yielded result", "principal", pk)
			respondDTO(w, result)
			return
		}

		slog.Info("Groups request can not be served properly", "token", auth)
		respondError(w, http.StatusNotFound, "groups not found")
	}
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

func main() {
	http.HandleFunc("/", notFoundHandler())
	http.HandleFunc("/api/v4/user", meHandler())
	http.HandleFunc("/api/v4/groups", groupsHandler())

	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("HTTP Server error", "err", err)
		os.Exit(1)
	}
}
