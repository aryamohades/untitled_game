package session

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"untitled_game/core/token"
)

// ErrInvalidAuthToken is used when the supplied auth token is formatted incorrectly.
var ErrInvalidAuthToken = errors.New("invalid auth token")

// tokenDelimiter is used to separate the user id and session key in the auth token.
const tokenDelimiter = ":"

// Session represents a user session.
type Session struct {
	ID  int
	Key string
}

// New creates a new session for account with the given id.
func New(id int) (Session, error) {
	key, err := token.Generate(32)
	if err != nil {
		return Session{}, err
	}
	return Session{id, key}, nil
}

// Token represents an auth token that is created following a successful authentication attempt.
type Token struct {
	Token string `json:"token"`
}

// CreateToken creates an auth token from a session.
func CreateToken(sess Session) Token {
	return Token{strconv.Itoa(sess.ID) + tokenDelimiter + sess.Key}
}

// ParseToken parses an auth token into a session.
func ParseToken(token string) (Session, error) {
	parts := strings.SplitN(token, tokenDelimiter, 2)
	if len(parts) != 2 {
		return Session{}, ErrInvalidAuthToken
	}

	id, err := strconv.Atoi(parts[0])
	if err != nil {
		return Session{}, ErrInvalidAuthToken
	}
	return Session{id, parts[1]}, nil
}

type contextKey int

const contextKeySession contextKey = iota

// WithSession adds the user session to the http request context.
func WithSession(r *http.Request, sess Session) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), contextKeySession, sess))
}

// GetSession retrieves the current session from the http request context.
func GetSession(r *http.Request) Session {
	return r.Context().Value(contextKeySession).(Session)
}
