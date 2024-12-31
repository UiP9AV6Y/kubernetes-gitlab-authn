package web

import (
	"log/slog"
	"net/http"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/gitlab-mock/internal/model"
)

func NotFoundHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		logger.Warn("Missing route implementation", "path", req.URL.Path)
		respondError(w, http.StatusNotFound, "not implemented")
	}
}

func MeHandler(q model.SelectUserQuery, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
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
}

func GroupsHandler(q model.SelectGroupsQuery, logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("Groups request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		result, ok := q(auth)
		if ok {
			logger.Info("Groups request yielded result", "token", auth)
			respondDTO(w, result)
			return
		}

		logger.Info("Groups request can not be served properly", "token", auth)
		respondError(w, http.StatusNotFound, "groups not found")
	}
}
