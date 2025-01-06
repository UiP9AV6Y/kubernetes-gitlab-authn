package web

import (
	"log/slog"
	"net/http"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
)

const (
	metadataVersion  = "0.0.0-mock"
	metadataRevision = "0000000"
)

func NotFoundHandler(logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		logger.Warn("Missing route implementation", "path", req.URL.Path)
		respondError(w, http.StatusNotFound, "not implemented")
	}

	return http.HandlerFunc(handler)
}

func VersionHandler(logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("Version request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		result := map[string]interface{}{
			"version":  metadataVersion,
			"revision": metadataRevision,
		}
		logger.Info("Version requested", "token", auth)
		respondDTO(w, result)
	}

	return http.HandlerFunc(handler)
}

func MetaDataHandler(logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("Metadata request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		kas := map[string]interface{}{
			"enabled": false,
			"version": metadataVersion,
		}
		result := map[string]interface{}{
			"version":    metadataVersion,
			"revision":   metadataRevision,
			"enterprise": false,
			"kas":        kas,
		}
		logger.Info("Metadata requested", "token", auth)
		respondDTO(w, result)
	}

	return http.HandlerFunc(handler)
}

func MeHandler(q model.SelectUserQuery, logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("User request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		result, ok := q(auth)
		if ok {
			logger.Info("User request yielded result", "token", auth)
			respondDTO(w, result)
			return
		}

		logger.Info("User request can not be served properly", "token", auth)
		respondError(w, http.StatusNotFound, "user not found")
	}

	return http.HandlerFunc(handler)
}

func GroupsHandler(q model.SelectGroupsQuery, logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("Groups request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		page, size := parseOffsetPagination(req)
		offset := size * (page - 1)
		result, total, ok := q(auth, offset, size)
		if ok {
			logger.Info("Groups request yielded result", "token", auth, "page", page, "per_page", size, "total", total)
			NewPagination(size, page, total).WriteHeader(w, req)
			respondDTO(w, result)
			return
		}

		logger.Info("Groups request can not be served properly", "token", auth, "page", page, "per_page", size)
		respondError(w, http.StatusNotFound, "groups not found")
	}

	return http.HandlerFunc(handler)
}
