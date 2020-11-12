package auth

import (
	"errors"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// ErrInvalidCredentials is used when authentication fails due to an incorrect account email and
// password combination being supplied.
var ErrInvalidCredentials = errors.New("invalid account credentials")

// Account represents account info that is retrieved from the account repository in order to
// complete the authentication process. The retrieved account info is compared to the supplied info
// to check for a match. If successful, the retrieved ID is associated with a generated session key
// in order to identify the user on subsequent requests to the server.
type Account struct {
	ID       int    `json:"id"`
	Password string `json:"password"`
}

// Credentials defines an email and password combination used to authenticate a user.
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates the Credentials fields.
func (c Credentials) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Email, validation.Required, is.Email),
		validation.Field(&c.Password, validation.Required),
	)
}
