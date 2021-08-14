package main

import (
	"fmt"
	"net/http"
)

// logError - log generic erros
func (app *application) logError(r *http.Request, err error) {
	app.logger.Println(err)
}


// errorResponse send JSON formatted error messages to the client with status code. 
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := envelope{
		"error": message,
	}

	// Write the response with writeJSON
	// If this fails fall back nd send the client an empty response with status 500
	err := app.writeJSON(w, status, env, nil) 
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// serverErrorResponse used when application encounters a problem at runtime
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"

	app.errorResponse(w, r, http.StatusInternalServerError, message)
} 

func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "the requested resource could not be found"
	app.errorResponse(w,r,http.StatusNotFound, message)
}

// methodNotAllowed sends a 405
func (app *application) methodNotAllowedResponse(w http.ResponseWriter, r *http.Request) {
	message := fmt.Sprintf("The %s method is not supported for this resource", r.Method)
	app.errorResponse(w,r,http.StatusMethodNotAllowed, message)
}

func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w,r, http.StatusBadRequest, err.Error())
}