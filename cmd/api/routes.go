package main

import (
	"github.com/julienschmidt/httprouter"
	"net/http"
)

func (app *application)routes() *httprouter.Router {
	// Initialize httprouter
	router := httprouter.New()

	// Register our routes
	// URL patterns and handler functions
	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthCheckHandler)
	router.HandlerFunc(http.MethodPost, "/v1/movies", app.createMovieHandler)
	router.HandlerFunc(http.MethodGet, "/v1/movies/:id", app.showMovieHandler)

	// return an httprouter.
	return router
}