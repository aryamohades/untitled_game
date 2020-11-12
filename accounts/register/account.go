package register

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

// NewAccount represents the data required to register a new account.
type NewAccount struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Validate validates new account data.
func (a NewAccount) Validate() error {
	return validation.ValidateStruct(&a,
		validation.Field(&a.Email, validation.Required, is.Email),
		validation.Field(&a.Password, validation.Required),
	)
}
