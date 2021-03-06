package main

import (
	"expvar"
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
	// This was before we checked for certain permissions
	// movie routes require an activated user.
	// router.HandlerFunc(http.MethodPost, "/v1/movies", app.requireActivatedUser(app.createMovieHandler))
	// router.HandlerFunc(http.MethodGet, "/v1/movies", app.requireActivatedUser(app.listMoviesHandler))
	// router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requireActivatedUser(app.showMovieHandler))
	// router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requireActivatedUser(app.updateMovieHandler))
	// router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requireActivatedUser(app.deleteMovieHandler))
	// router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.requiredActivatedUser(app.updateMovieHandler))
	router.HandlerFunc(http.MethodGet, "/v1/movies", app.requirePermission("movies:read",app.listMoviesHandler))
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.requirePermission("movies:write",app.createMovieHandler))
	
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.requirePermission("movies:read",app.showMovieHandler))
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.requirePermission("movies:write",app.updateMovieHandler))
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.requirePermission("movies:write",app.deleteMovieHandler))

	// Route to create our user
	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)

	// Token route
	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)

	// Register a new GET enpoint that will display data
	router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())

	// return an httprouter.
	return app.metrics(app.recoverPanic(app.enableCors(app.rateLimit(app.authenticate(router)))))
}
