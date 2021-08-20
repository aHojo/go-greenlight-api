package main

import (
	"errors"
	"net/http"
	"time"

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
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	// After the user record has been created in the database, generate a new activation token for the user
	token, err := app.models.Token.New(user.ID, 3*24*time.Hour, data.ScopeActivation)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// use the background helper to send the emails using an anon function.
	app.background(func() {
		// AS the is now multiple pieces of data that need to be sent to the template
		// we create a holding struct for the data
		data := map[string]interface{}{
			"activationToken": token.Plaintext,
			"userID":          user.ID,
		}
		// We are using a third party sender, and a panic can happen. We will have to recover from it.
		// Call the send method to send the emailcom
		err = app.mailer.Send(user.Email, "user_welcome.gotmpl", data)
		if err != nil {
			// Important - if there is an error sending the email then we us the
			// app.logger.PrintError() helper to manage it
			// Because we do not want to send a status code. This code will complete after the handler.
			app.logger.PrintError(err, nil)
		}
	})

	// Write the json response
	err = app.writeJSON(w, http.StatusCreated, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}

}

func (app *application) activateUserHandler(w http.ResponseWriter, r *http.Request) {

	// Parse the plaintext activation token from the body
	var input struct {
		TokenPlaintext string `json:"token"`
	}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	// Validate the plaintext token provided by the client
	v := validator.New()

	// Retrieve the details of the user associated with the token, using the GetForToken() method.
	// If no matching record is found, then we let the client know the token is invalid
	user, err := app.models.Users.GetForToken(data.ScopeActivation, input.TokenPlaintext)
	if err != nil {
		switch{
		case errors.Is(err, data.ErrRecordNotFound):
			v.AddError("token", "invalid or expired activation token")
			app.failedValidationResponse(w,r,v.Errors)
		default: 
			app.serverErrorResponse(w,r,err)
		}
		return
	}


	// Update the user's activation status
	user.Activated = true

	// Save the updated user record in our database, checking for any edit conflicts in the same way we did for the movies
	err = app.models.Users.Update(user)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrEditConflict):
			app.ErrEditConflictResponse(w,r)
		default:
			app.serverErrorResponse(w,r, err)
		}
		return
	}

	// If everything went successfully, then we delte all activation token for the user
	err = app.models.Token.DeleteAllForUser(data.ScopeActivation, user.ID)
	if err != nil {
		app.serverErrorResponse(w,r,err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w,r,err)
	}
}
