package api

import "net/http"

// ErrInternalError is used when an unexpected error occurs. For unexpected errors, there is no
// specific action that the consumer can take to resolve the issue. Some unexpected errors are
// temporary, and trying the request again may resolve the issue.
var ErrInternalError = Error{Message: "Internal error.", Status: http.StatusInternalServerError}

// ErrRouteNotFound is used when a request is made to a route that is not explicitly handled by any
// http handler.
var ErrRouteNotFound = Error{Message: "Route not found.", Status: http.StatusNotFound}

// ErrMethodNotAllowed is used when a request is made using a request method that is not supported
// for that specific route.
var ErrMethodNotAllowed = Error{Message: "Method not allowed.", Status: http.StatusMethodNotAllowed}

// ErrContentType is used when a request is made with the incorrect Content-Type header. For POST
// equests that require a request body, the only accepted Content-Type header is application/json.
var ErrContentType = Error{Message: "Invalid Content-Type header.", Status: http.StatusUnsupportedMediaType}

// ErrInvalidRequestBody is used when a request is made with a request body that is formatted
// incorrectly. Some examples of incorrect request body formatting include being empty, having
// invalid JSON, or having mismatched types.
var ErrInvalidRequestBody = Error{Message: "Invalid request body.", Status: http.StatusBadRequest}

// ErrUnauthorized is used when a request is not authenticated or is not authorized to make the
// request or interact with a particular resource.
var ErrUnauthorized = Error{Message: "Unauthorized.", Status: http.StatusUnauthorized}

// ErrInvalidAuthToken is sent as an http response when the supplied auth token is invalid.
var ErrInvalidAuthToken = Error{Message: "Invalid auth token.", Status: http.StatusUnauthorized}

// ErrValidationError is used when the request body is formatted correctly, but one or more of the
// fields does not meet some requirement. An example is this is requiring a minimum length on a
// particular field.
var ErrValidationError = Error{Message: "There were some validation errors.", Status: http.StatusBadRequest}

// Error represents a custom error to be used as the response to an http request.
type Error struct {
	Message string      `json:"message"`
	Status  int         `json:"-"`
	Details interface{} `json:"details,omitempty"`
}

// WithDetails adds additional details to the error.
func (e Error) WithDetails(details interface{}) Error {
	e.Details = details
	return e
}

// Error implements the error interface and returns the error message.
func (e Error) Error() string {
	return e.Message
}
