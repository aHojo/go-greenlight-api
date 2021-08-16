package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ahojo/greenlight/internal/data"
	"github.com/ahojo/greenlight/internal/validator"
)

// createMovieHandler for "POST" /v1/movies
func (app *application) createMovieHandler(w http.ResponseWriter, r *http.Request) {

	// Declare an anonymous struct to hold the information we are expecting to be in the request body.
	var input struct {
		Title   string       `json:"title"`
		Year    int32        `json:"year"`
		Runtime data.Runtime `json:"runtime"`
		Genres  []string     `json:"genres"`
	}

	// use Jsone Decoder to read the request body,
	// Decode it
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Start of validation
	v := validator.New()
	movie := &data.Movie{
		Title: input.Title,
		Year: input.Year,
		Runtime: input.Runtime,
		Genres: input.Genres,
	}

	data.ValidateMovie(v, movie)

	// Use the valid method to see if any of the checks failed. If they did, then use 
	// failedValidationResponse() helper to send a response to the client
	if !v.Valid() {
		app.failedValidationResponse(w,r,v.Errors)
		return
	}

	// Dump the contents of the input struct in a HTTP response
	fmt.Fprintf(w, "%v\n", input)
}

// showMovieHandler for "GET /v1/movies/:id"
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Creaet a new instance of the movie struct
	movie := data.Movie{
		ID:        id,
		CreatedAt: time.Now(),
		Title:     "Rambo",
		Runtime:   102,
		Genres:    []string{"action", "romance", "war"},
		Version:   1,
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
