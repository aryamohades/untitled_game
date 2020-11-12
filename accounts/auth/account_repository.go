package auth

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
)

// ErrAccountNotFound is used when an account could not be found in the account repository.
var ErrAccountNotFound = errors.New("account not found")

// AccountRepository provides methods for interacting with an account store.
type AccountRepository interface {
	GetByEmail(email string) (Account, error)
}

type accountRepository struct {
	db *sqlx.DB
}

// NewAccountRepository creates a new postgres account repository.
func NewAccountRepository(db *sqlx.DB) AccountRepository {
	return &accountRepository{db}
}

// GetByEmail retrieves an account from the database by its email.
func (r *accountRepository) GetByEmail(email string) (Account, error) {
	const q = `SELECT id, password FROM accounts WHERE email = $1`

	var account Account
	if err := r.db.Get(&account, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return account, ErrAccountNotFound
		}
		return account, err
	}
	return account, nil
}
