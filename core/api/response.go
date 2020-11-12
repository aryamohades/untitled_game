package api

import (
	"log"
	"net/http"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Responder provides methods for responding to an http request.
type Responder interface {
	Respond(w http.ResponseWriter, data interface{})
	RespondError(w http.ResponseWriter, err error)
	RespondStatus(w http.ResponseWriter, code int)
}

type responder struct {
	log *log.Logger
}

// NewResponder creates a new responder.
func NewResponder(log *log.Logger) Responder {
	return &responder{log}
}

// Respond responds to the http request with some data. The data sent is assumed to be in JSON
// format. If it is not in JSON format, then a generic internal server error is sent. The status
// code of the response is not explicitly and is set to 200 OK by default.
func (r *responder) Respond(w http.ResponseWriter, data interface{}) {
	bytes, err := json.Marshal(data)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	r.checkErr(w.Write(bytes))
}

// RespondError responds to the http request with the provided error. If the error is a custom
// error, then it is marshalled and sent as-is as the response body. If the error is not a known
// error, then a generic internal server error response is sent.
func (r *responder) RespondError(w http.ResponseWriter, err error) {
	e, ok := err.(Error)
	if !ok {
		// An unexpected error has occured; log it and send back a generic internal server error.
		r.log.Printf("unexpected error: %v", err)
		e = ErrInternalError
	}

	bytes, err := json.Marshal(e)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(e.Status)
	r.checkErr(w.Write(bytes))
}

// RespondStatus sends an http response header with the provided status code. This is used in cases
// where a status code is a sufficient response and no response body is required.
func (r *responder) RespondStatus(w http.ResponseWriter, code int) {
	w.WriteHeader(code)
}

func (r *responder) checkErr(bytes int, err error) {
	if err != nil {
		r.log.Printf("write response: %v", err)
	}
}
