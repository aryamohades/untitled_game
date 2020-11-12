package register

import (
	"strings"
	"untitled_game/core/token"

	"golang.org/x/crypto/bcrypt"
)

// Service provides account registration related services.
type Service interface {
	CreateAccount(account NewAccount) error
}

type service struct {
	accounts AccountRepository
}

// NewService creates a new account registration service.
func NewService(accounts AccountRepository) Service {
	return &service{accounts}
}

// CreateAccount creates a new account.
func (s *service) CreateAccount(account NewAccount) error {
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(account.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	token, err := token.Generate(32)
	if err != nil {
		return err
	}

	account.Email = strings.ToLower(account.Email)
	account.Password = string(hashedPw)

	return s.accounts.Create(account, token)
}
