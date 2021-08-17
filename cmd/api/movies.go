package main

import (
	"errors"
	"fmt"
	"net/http"

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
		Title:   input.Title,
		Year:    input.Year,
		Runtime: input.Runtime,
		Genres:  input.Genres,
	}

	data.ValidateMovie(v, movie)

	// Use the valid method to see if any of the checks failed. If they did, then use
	// failedValidationResponse() helper to send a response to the client
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Call the Insert() method on our model.
	// Creates a record in the database, and update the movie struct passed in with the
	// system generated info
	err = app.models.Movies.Insert(movie)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// When sending an HTTP response, we want to include a Location header to let the
	// client know which URL they can find the new-resource at.
	headers := make(http.Header)
	headers.Set("Location", fmt.Sprintf("/v1/movie/%d", movie.ID))

	// Write a json response with a 201 Created
	err = app.writeJSON(w, http.StatusCreated, envelope{"movie": movie}, headers)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

// listMoviesHandler will list multiple movies and allow us to paginate, sort, and filter.
func (app *application) listMoviesHandler(w http.ResponseWriter, r *http.Request) {
	// input struct to hold expected values
	var input struct {
		Title   string
		Genres  []string
		Filters 	data.Filters
	}

	// Initialize a validator
	v := validator.New()

	// Call r.Url.Query to get the url.Values map containing the Query string data.
	qs := r.URL.Query()

	// Use the helper functions we defined to get the title and genres query string values
	// Fall back to defaults of an empty string and an empty string slice
	input.Title = app.readString(qs, "title", "")
	input.Genres = app.readCSV(qs, "genres", []string{})

	// Get the page and page_size query string values as integers. Notice that we set
	// the default page value to 1 and default page_size to 20, and that we pass the
	// validator instance as the final argument here
	input.Filters.Page = app.readInt(qs, "page", 1, v)
	input.Filters.PageSize = app.readInt(qs, "page_size", 20, v)

	// Extract the sort query  string value, falling back to "id" if it is not provided
	input.Filters.Sort = app.readString(qs, "sort", "id")
	// list of things we are allowed to sort on.
	input.Filters.SortSafelist = []string{
		"id",
		"title",
		"year",
		"runtime",
		"-id",
		"-title",
		"-year",
		"-runtime",
	}
	
	// Execute the vailadtion checks on the Filters struct
	// Check the validator
	if data.ValidateFilters(v, input.Filters); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Dump the contents of the input struct in a http resposne
	fmt.Fprintf(w, "%+v\n", input)
}

// showMovieHandler for "GET /v1/movies/:id"
func (app *application) showMovieHandler(w http.ResponseWriter, r *http.Request) {

	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Call the Get() method to fetch the data for a specific movie. We also need to
	// use the errors.Is() function to check if it returns a data.ErrRecordNotFound
	// error, in which case we send a 404 Not Found response to the client.
	movie, err := app.models.Movies.Get(id)

	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) updateMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Get the id from the url.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
	}

	// Fetch the existing movie record from the database
	// 404 if not found
	movie, err := app.models.Movies.Get(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// input struct for the expected values from the client
	// NOTE: pointers have a nil zero value.
	var input struct {
		Title   *string       `json:"title"`   // Will be nil if not given
		Year    *int32        `json:"year"`    // Will be nil if not given
		Runtime *data.Runtime `json:"runtime"` // Will be nil if not given
		Genres  []string      `json:"genres"`  // slice already has a nil zero value
	}

	// Read the JSON
	err = app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
	}

	// Copy the values from the request body to the movie struct
	// If these are nil we know that it was not given in the request body.
	if input.Title != nil {
		movie.Title = *input.Title
	}
	if input.Year != nil {
		movie.Year = *input.Year
	}
	if input.Runtime != nil {
		movie.Runtime = *input.Runtime
	}
	if input.Genres != nil {
		movie.Genres = input.Genres
	}

	// Validation that the data is ok.
	v := validator.New()
	if data.ValidateMovie(v, movie); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Update the movie after verified.
	err = app.models.Movies.Update(movie)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.ErrEditConflictResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"movie": movie}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) deleteMovieHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the movie ID from the URL.
	id, err := app.readIDParam(r)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Delete the movie from the database, sending a 404 Not Found response to the
	// client if there isn't a matching record.
	err = app.models.Movies.Delete(id)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.notFoundResponse(w, r)
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	// Return a 200 OK status code along with a success message.
	err = app.writeJSON(w, http.StatusOK, envelope{"message": "movie successfully deleted"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
