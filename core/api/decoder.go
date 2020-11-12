package api

import (
	"io"
	"net/http"
)

// Decoder provides a method Decode for parsing a JSON request body.
type Decoder interface {
	Decode(w http.ResponseWriter, r *http.Request, dest interface{}) error
}

// DecoderConfig represents configuration options for a decoder.
type DecoderConfig struct {
	MaxBytes int64
}

// StandardDecoderConfig represents sane default configuration for a decoder.
var StandardDecoderConfig = DecoderConfig{
	MaxBytes: 1048576, // 1MB
}

type decoder struct {
	maxBytes int64
}

// NewDecoder creates a new decoder.
func NewDecoder(cfg DecoderConfig) Decoder {
	if cfg.MaxBytes == 0 {
		cfg.MaxBytes = StandardDecoderConfig.MaxBytes
	}
	return &decoder{
		maxBytes: cfg.MaxBytes,
	}
}

// Decode attempts to parse a JSON request body into the destination type.
func (d *decoder) Decode(w http.ResponseWriter, r *http.Request, dest interface{}) error {
	// If the Content-Type header is present, check that it has the value application/json. If the
	// header is present and is not application/json, then return a content type error.
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		return ErrContentType
	}

	// Use http.MaxBytesReader to enforce a limit on the number of bytes read from the request
	// body. If the request body exceeds the maximum number of bytes allowed, then Decode will
	// return an error.
	r.Body = http.MaxBytesReader(w, r.Body, d.maxBytes)

	// Initialize the decoder and call the DisallowUnknownFields() method on it. This will cause
	// Decode to return an error if it encounters any unexpected fields in the JSON data.
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the request body and capture the potential error. Reasons why Decode  may fail
	// include poorly formed JSON, field type mismatches in the destination type, unexpected
	// fields in the JSON data, empty request body, or a request body that exceeds the maximum
	// number of bytes allowed.
	err := dec.Decode(&dest)
	if err != nil {
		return ErrInvalidRequestBody
	}

	// Decode does not read the entire request body in one read, so it is possible to have
	// additional data following our valid JSON data. Call Decode again to check for additional
	// data. If this results in an error and the error is not io.EOF, then there was additional
	// data in the request body, and an invalid request body error is returned.
	if err := dec.Decode(&struct{}{}); err != nil && err != io.EOF {
		return ErrInvalidRequestBody
	}
	return nil
}
