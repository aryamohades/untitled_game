package migrate

import (
	"errors"
	"fmt"
	"net/url"
	"untitled_game/core/postgres"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

// Migrate applies all database migrations.
func Migrate(databaseURL string) error {
	if err := ensureDatabase(databaseURL); err != nil {
		return err
	}

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}
	return nil
}

func ensureDatabase(databaseURL string) error {
	parsedURL, err := url.Parse(databaseURL)
	if err != nil {
		return err
	}

	if len(parsedURL.Path) == 0 {
		return errors.New("no path specified in database url")
	}

	dbname := parsedURL.Path[1:]
	parsedURL.Path = ""

	db, err := sqlx.Open("postgres", parsedURL.String())
	if err != nil {
		return err
	}
	defer db.Close()

	if err := postgres.Status(db); err != nil {
		return err
	}

	if _, err := db.Exec(fmt.Sprintf("CREATE DATABASE %q", dbname)); err != nil && !postgres.IsDuplicateDatabaseError(err) {
		return err
	}
	return nil
}
