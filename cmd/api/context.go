package main

import (
	"context"
	"net/http"

	"github.com/ahojo/greenlight/internal/data"
)

// Define a custom contextKey type, with the underlying type string
type contextKey string

// convert the string "user" to a contextKey type and assign it to the userContextKey constant
// We'll use this constant as the key for getting and setting user information in the request context
const userContextKey = contextKey("user")

// contextSetUser() method returns a new copy of the request with the provided User Struct added to the context
// NOTE: userContextKey as the key
func (app *application) contextSetUser(r *http.Request, user *data.User) *http.Request {

	ctx	 := context.WithValue(r.Context(), userContextKey, user)
	return r.WithContext(ctx)
}

// contxtGetUser retrieves the User struct from the request context.
// The only time that we'll use this is when we logically expect there to be User struct 
// value in the context. If it doesn't exist, it will be an "unexecpeted error"
func (app *application) contextGetUser(r *http.Request) *data.User {
	user, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing user value in request context")
	}
	return user
}

