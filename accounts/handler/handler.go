package handler

import (
	"log"
	"net/http"
	"untitled_game/accounts/auth"
	"untitled_game/accounts/middleware"
	"untitled_game/accounts/register"
	"untitled_game/accounts/session"
	"untitled_game/core/api"
)

// New creates a new http handler and attaches routes.
func New(log *log.Logger, sess session.Store, authService auth.Service, registerService register.Service) http.Handler {
	dec := api.NewDecoder(api.StandardDecoderConfig)
	res := api.NewResponder(log)
	h := api.NewHandler(log, res)

	authMw := middleware.Authenticate(res, sess)

	authHandler := &authHandler{dec, res, authService}
	h.Handle(http.MethodGet, "/session", authHandler.getSession, authMw)
	h.Handle(http.MethodPost, "/session", authHandler.createSession)
	h.Handle(http.MethodDelete, "/session", authHandler.deleteSession, authMw)
	h.Handle(http.MethodPost, "/authenticate", authHandler.authenticate)

	registerHandler := &registerHandler{dec, res, registerService}
	h.Handle(http.MethodPost, "/register", registerHandler.registerAccount)

	return h
}
