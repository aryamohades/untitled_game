package postgres

import (
	"net/url"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	uniqueViolationErrorCode   = "23505"
	duplicateDatabaseErrorCode = "42P04"
)

// Config represents configuration options for a postgres connection pool.
type Config struct {
	Address         string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifetime time.Duration
}

// Open creates a postgres connection pool with the provided configuration.
func Open(cfg Config) (*sqlx.DB, error) {
	parsedURL, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	query := parsedURL.Query()
	if len(query.Get("sslmode")) == 0 {
		query.Set("sslmode", "disable")
		parsedURL.RawQuery = query.Encode()
	}

	db, err := sqlx.Open("postgres", parsedURL.String())
	if err != nil {
		return nil, err
	}

	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)

	return db, nil
}

// Status executes a query against the database to determine if the connection is valid.
func Status(db *sqlx.DB) error {
	q := `SELECT true`
	var status bool
	return db.QueryRow(q).Scan(&status)
}

// Version queries the database version from postgres.
func Version(db *sqlx.DB) (string, error) {
	var version string
	err := db.QueryRow("SELECT version()").Scan(&version)
	return version, err
}

// IsUniqueViolationError returns true if the error is a postgres unique violation error.
func IsUniqueViolationError(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Code == uniqueViolationErrorCode
}

// IsDuplicateDatabaseError returns true if the error is a postgres duplicate database error.
func IsDuplicateDatabaseError(err error) bool {
	pgErr, ok := err.(*pq.Error)
	return ok && pgErr.Code == duplicateDatabaseErrorCode
}
