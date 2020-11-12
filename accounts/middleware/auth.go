package middleware

import (
	"net/http"
	"strings"
	"untitled_game/accounts/session"
	"untitled_game/core/api"
)

type sessionStore interface {
	Get(sess session.Session) (session.Session, error)
}

// Authenticate validates a session token from the Authorization header of the request. If the
// session store contains an entry for the provided token, then the token is considered valid
// and the session is added to the request context. If the token is invalid, then the middleware
// responds to the request with an unauthorized error.
func Authenticate(res api.Responder, sessions sessionStore) api.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			parts := strings.SplitN(r.Header.Get("Authorization"), " ", 2)
			if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
				res.RespondError(w, api.ErrUnauthorized)
				return
			}

			parsedSess, err := session.ParseToken(parts[1])
			if err != nil {
				res.RespondError(w, api.ErrInvalidAuthToken)
				return
			}

			sess, err := sessions.Get(parsedSess)
			if err != nil {
				res.RespondError(w, api.ErrUnauthorized)
				return
			}
			next.ServeHTTP(w, session.WithSession(r, sess))
		}
	}
}
