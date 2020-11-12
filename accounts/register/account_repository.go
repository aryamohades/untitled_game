package register

import (
	"errors"
	"untitled_game/core/postgres"

	"github.com/jmoiron/sqlx"
)

// ErrAccountExists is used when an account cannot be created due to the supplied email address
// already being in use.
var ErrAccountExists = errors.New("account already exists")

// AccountRepository provides methods for interacting with an account store.
type AccountRepository interface {
	Create(account NewAccount, token string) error
}

type accountRepository struct {
	db *sqlx.DB
}

// NewAccountRepository creates a new postgres account repository.
func NewAccountRepository(db *sqlx.DB) AccountRepository {
	return &accountRepository{db}
}

// Creates inserts a new account into the database.
func (r *accountRepository) Create(account NewAccount, token string) error {
	const q = `INSERT INTO accounts (email, password, verification_token, verification_token_expires_at) VALUES ($1, $2, $3, now() + interval '1 day')`

	if _, err := r.db.Exec(q, account.Email, account.Password, token); err != nil {
		if postgres.IsUniqueViolationError(err) {
			return ErrAccountExists
		}
		return err
	}
	return nil
}
