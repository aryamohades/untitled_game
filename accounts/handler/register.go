package handler

import (
	"errors"
	"net/http"
	"untitled_game/accounts/register"
	"untitled_game/core/api"
)

// errAccountExists is sent as an http response when the user attempts to create an account with an
// email address that is already in use.
var errAccountExists = api.Error{Message: "Account already exists", Status: http.StatusConflict}

type registerHandler struct {
	dec api.Decoder
	res api.Responder
	s   register.Service
}

func (h *registerHandler) registerAccount(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var account register.NewAccount
	if err := h.dec.Decode(w, r, &account); err != nil {
		h.res.RespondError(w, err)
		return
	}

	if err := account.Validate(); err != nil {
		h.res.RespondError(w, api.ErrValidationError.WithDetails(err))
		return
	}

	if err := h.s.CreateAccount(account); err != nil {
		if errors.Is(err, register.ErrAccountExists) {
			h.res.RespondError(w, errAccountExists)
			return
		}
		h.res.RespondError(w, err)
		return
	}
	h.res.RespondStatus(w, http.StatusCreated)
}
