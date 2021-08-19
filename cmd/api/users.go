package main

import (
	"errors"
	"net/http"

	"github.com/ahojo/greenlight/internal/data"
	"github.com/ahojo/greenlight/internal/validator"
)

func (app *application) registerUserHandler(w http.ResponseWriter, r *http.Request) {

	// Create an anonymous struct to hold the expected data from the request body.
	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse the request body into the anonymous struct
	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Copy the data from the request body into a new User struct. Notice also that we
	// set the Activated field to false, which isn't strictly necessary because the
	// Activated field will have the zero-value of false by default. But setting this
	// explicitly helps to make our intentions clear to anyone reading the code.
	user := &data.User{
		Name:      input.Name,
		Email:     input.Email,
		Activated: false,
	}

	// Use the password.set method to genreate and store the hashed password and plaintext password.
	err = user.Password.Set(input.Password)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	//  New instance of our validator
	v := validator.New()

	// Validate the user struct and return the error messages to the client if any fail
	if data.ValidateUser(v, user); !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the user data into the database
	err = app.models.Users.Insert(user)
	if err != nil {
		switch {
		// If we get a duplicate email error, us the v.AddError()
		case errors.Is(err, data.ErrDuplicateEmail):
			v.AddError("email", "a user with this email already exists")
			app.failedValidationResponse(w, r, v.Errors)
		default:
			app.serverErrorResponse(w,r,err)
		}
		return
	}

	// Write the json response 
	err = app.writeJSON(w,http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w,r,err)
	}
}
