package auth

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// Account represents account info that is retrieved from the account repository as part of the
// authentication process. The retrieved account password is compared to the supplied password
// using bcrypt to check for a match.
type Account struct {
	ID       int    `json:"id"`
	Password string `json:"password"`
}

// Credentials represents an email and password combination that is used to authenticate a user.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates account credentials data.
func (c Credentials) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required, is.Email),
		validation.Field(&c.Password, validation.Required),
	)
}
