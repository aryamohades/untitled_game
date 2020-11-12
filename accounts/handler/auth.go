package handler

import (
	"errors"
	"net/http"
	"untitled_game/accounts/auth"
	"untitled_game/accounts/session"
	"untitled_game/core/api"
)

// errInvalidCredentials is sent as an http response when authentication fails due to an incorrect
// account email address and password combination being supplied by the user.
var errInvalidCredentials = api.Error{Message: "Invalid account credentials.", Status: http.StatusUnauthorized}

type authHandler struct {
	dec api.Decoder
	res api.Responder
	s   auth.Service
}

func (h *authHandler) getSession(w http.ResponseWriter, r *http.Request) {
	h.res.RespondStatus(w, http.StatusOK)
}

func (h *authHandler) createSession(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var creds auth.Credentials
	if err := h.dec.Decode(w, r, &creds); err != nil {
		h.res.RespondError(w, err)
		return
	}

	if err := creds.Validate(); err != nil {
		h.res.RespondError(w, api.ErrValidationError.WithDetails(err))
		return
	}

	token, err := h.s.Login(creds)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			h.res.RespondError(w, errInvalidCredentials)
			return
		}
		h.res.RespondError(w, err)
		return
	}
	h.res.Respond(w, token)
}

func (h *authHandler) deleteSession(w http.ResponseWriter, r *http.Request) {
	sess := session.GetSession(r)
	if err := h.s.Logout(sess); err != nil {
		if errors.Is(err, session.ErrSessionNotFound) {
			h.res.RespondError(w, api.ErrUnauthorized)
			return
		}
		h.res.RespondError(w, err)
		return
	}
	h.res.RespondStatus(w, http.StatusOK)
}
