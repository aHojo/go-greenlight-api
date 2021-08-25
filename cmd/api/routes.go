package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

// Update to return http.Handler instead of *httprouter.Router because of middleware.
func (app *application) routes() http.Handler {
	// Initialize httprouter
	router := httprouter.New()

	// Conver the notFoundResponse() helpper to a http.Handler and set it as
	// the custom error handler for 404
	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	// Register our routes
	// URL patterns and handler functions
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.listMoviesHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	// router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// Route to create our user
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// Token route
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// return an httprouter.
	return app.recoverPanic(app.rateLimit(app.authenticate(router)))
}
