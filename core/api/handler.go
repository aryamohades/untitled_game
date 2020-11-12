package api

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/julienschmidt/httprouter"
)

// Handler represents an http handler which is comprised of an http router and core middleware to
// execute for each handler route.
type Handler struct {
	router *httprouter.Router
	mw     []Middleware
}

// NewHandler creates a new http handler with standard configuration.
func NewHandler(log *log.Logger, res Responder, mw ...Middleware) *Handler {
	router := httprouter.New()

	router.PanicHandler = func(w http.ResponseWriter, r *http.Request, v interface{}) {
		// Log the panic and stack trace.
		log.Printf("panic: %v", v)
		log.Println(string(debug.Stack()))

		// Respond with a generic internal server error.
		res.RespondError(w, ErrInternalError)
	}

	router.NotFound = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res.RespondError(w, ErrRouteNotFound)
	})

	router.MethodNotAllowed = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		res.RespondError(w, ErrMethodNotAllowed)
	})

	// Turn off trailing slash redirection to require exact route matching.
	router.RedirectTrailingSlash = false

	return &Handler{router, mw}
}

// Handle adds an http request handler to the router for a specific route and request method. The
// provided handler function is wrapped in all of the provided middleware, as well as any core
// handler middleware.
func (h *Handler) Handle(method string, path string, handler http.HandlerFunc, mw ...Middleware) {
	// Wrap route specific middleware around this handler.
	handler = wrapMiddleware(mw, handler)

	// Add core middleware to the handler chain.
	handler = wrapMiddleware(h.mw, handler)

	// Add handler to app router.
	h.router.HandlerFunc(method, path, handler)
}

// ServeHTTP implements the http.Handler interface.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.router.ServeHTTP(w, r)
}
