package auth

import (
	"errors"
	"strings"
	"untitled_game/accounts/session"

	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidCredentials is used when authentication fails due to an incorrect account email
// address and password combination being supplied.
var ErrInvalidCredentials = errors.New("invalid account credentials")

// Service provides authentication related services.
type Service interface {
	Login(creds Credentials) (session.Token, error)
	Logout(sess session.Session) error
	Authenticate(creds Credentials) (session.Token, error)
}

type service struct {
	sess     session.Store
	accounts AccountRepository
}

// NewService creates a new auth service.
func NewService(sess session.Store, accounts AccountRepository) Service {
	return &service{sess, accounts}
}

// Login authenticates account credentials. If successful, a new session is added to the session
// store for the authenticated user.
func (s *service) Login(creds Credentials) (session.Token, error) {
	account, err := s.accounts.GetByEmail(strings.ToLower(creds.Email))
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return session.Token{}, ErrInvalidCredentials
		}
		return session.Token{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(creds.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return session.Token{}, ErrInvalidCredentials
		}
		return session.Token{}, err
	}

	sess, err := session.New(account.ID)
	if err != nil {
		return session.Token{}, err
	}

	if err := s.sess.Add(sess); err != nil {
		return session.Token{}, err
	}
	return session.CreateToken(sess), nil
}

// Logout logs the user out of the current session by deleting the session from the session store.
func (s *service) Logout(sess session.Session) error {
	return s.sess.Remove(sess)
}

// Authenticate authenticates account credentials. If successful a new session is returned for the
// authenticated user without adding the session to the session store.
func (s *service) Authenticate(creds Credentials) (session.Token, error) {
	account, err := s.accounts.GetByEmail(strings.ToLower(creds.Email))
	if err != nil {
		if errors.Is(err, ErrAccountNotFound) {
			return session.Token{}, ErrInvalidCredentials
		}
		return session.Token{}, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(creds.Password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return session.Token{}, ErrInvalidCredentials
		}
		return session.Token{}, err
	}

	sess, err := session.New(account.ID)
	if err != nil {
		return session.Token{}, err
	}
	return session.CreateToken(sess), nil
}
