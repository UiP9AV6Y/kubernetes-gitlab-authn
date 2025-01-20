package web

import (
	"errors"
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

func MeHandler(dao *model.DataAccess, logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("User request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		uid, err := dao.Tokens.FindUserIdentifier(auth)
		if errors.Is(err, model.ErrNotFound) {
			logger.Info("User request with invalid authentication", "token", auth)
			respondError(w, http.StatusNotFound, "user not found")
		} else if err != nil {
			logger.Info("User request can not be served properly", "err", err)
			respondError(w, http.StatusInternalServerError, "token lookup failed")
		}

		result, err := dao.Users.FindByIdentifier(uid)
		if err != nil {
			logger.Info("User request yielded no result", "uid", uid, "err", err)
			respondError(w, http.StatusInternalServerError, "user lookup failed")
		}

		logger.Info("User request yielded result", "uid", uid)
		respondDTO(w, result)
	}

	return http.HandlerFunc(handler)
}

func GroupsHandler(dao *model.DataAccess, logger *slog.Logger) http.Handler {
	handler := func(w http.ResponseWriter, req *http.Request) {
		auth := parseToken(req)
		if auth == "" {
			logger.Info("Groups request is missing authentication information")
			respondError(w, http.StatusUnauthorized, "missing token")
			return
		}

		uid, err := dao.Tokens.FindUserIdentifier(auth)
		if errors.Is(err, model.ErrNotFound) {
			logger.Info("Groups request with invalid authentication", "token", auth)
			respondError(w, http.StatusNotFound, "groups not found")
		} else if err != nil {
			logger.Info("Groups request can not be served properly", "err", err)
			respondError(w, http.StatusInternalServerError, "token lookup failed")
		}

		page, size := parseOffsetPagination(req)
		offset := size * (page - 1)

		result, err := dao.Groups.FindByUserIdentifier(uid, offset, size)
		if err != nil {
			logger.Info("Groups request yielded no result", "uid", uid, "err", err)
			respondError(w, http.StatusInternalServerError, "groups lookup failed")
		}

		total, err := dao.Groups.CountByUserIdentifier(uid)
		if err != nil {
			logger.Info("Unable to count user groups", "uid", uid, "err", err)
			respondError(w, http.StatusInternalServerError, "groups total failed")
		}

		logger.Info("Groups request yielded result", "uid", uid, "page", page, "per_page", size, "total", total)
		NewPagination(size, page, total).WriteHeader(w, req)
		respondDTO(w, result)
	}

	return http.HandlerFunc(handler)
}
