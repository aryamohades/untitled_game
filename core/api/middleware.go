package api

import "net/http"

// Middleware represents a handler function to run inbetween the request and its designated final
// handler function. Middleware can be used for authenticating or authorizing a request, adding
// additional request context, logging, and more.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// wrapMiddleware creates a new handler by wrapping middleware around a final handler. Each handler
// in the chain will be executed in the order they are provided.
func wrapMiddleware(mw []Middleware, handler http.HandlerFunc) http.HandlerFunc {
	// Iterate backwards through the middleware to ensure that the first middleware in the slice is
	// the first middleware to be executed.
	for i := len(mw) - 1; i >= 0; i-- {
		h := mw[i]
		if h != nil {
			handler = h(handler)
		}
	}
	return handler
}
