package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application)routes() *httprouter.Router {
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
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)
	// router.HandlerFunc(http.MethodPut, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodPatch, "/v1/movies/:id", app.updateMovieHandler)
	router.HandlerFunc(http.MethodDelete, "/v1/movies/:id", app.deleteMovieHandler)

	// return an httprouter.
	return router
}