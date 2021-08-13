package main

import (
	"errors"
	"github.com/julienschmidt/httprouter"
	"net/http"
	"strconv"
)

// readIDParam - gets the ID URL parameter from the current context
func (app *application) readIDParam(r *http.Request) (int64, error){
	// When httprouter parses a request, interpolated parameters will be stored
	// in the request context. Use ParamsFromContext() function to
	// get the slice containing them.
	params := httprouter.ParamsFromContext(r.Context())

	// Use the ByName() method to get the value of the id parameter from the slice
	// ByName() always returns a string, se we will convert it to a base 10 int
	// if the id is invalid return 404
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}
