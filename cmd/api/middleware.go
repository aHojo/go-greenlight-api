package main

import (
	"fmt"
	"net/http"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Create a deferred function (which will always be run in the event of a panic
		// as Go unwinds the stack
		defer func() {
			// Builtin recover function to check if there has been a panic or not
			if err := recover(); err != nil {
				// If there was a panic, set a "Connection: close" header
				w.Header().Set("Connection", "close")

				// Value return by recover is interface{}
				// fmt.Errorf to normalize it into an error and call our
				// serverErrorResponse() helper.
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}
