package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/ahojo/greenlight/internal/data"
	"github.com/ahojo/greenlight/internal/validator"
	"golang.org/x/time/rate"
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

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}

	// Declare a mutex and a map to hold IP Addresses
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	// Launch a background goroutine which removes old entries every minute
	go func() {
		for {
			time.Sleep(time.Minute)

			// Lock the mutex to prevent any rate limiter check from happening while cleanup is happening
			mu.Lock()

			// Loop through all the clients. If they haven't been seen within the last three minutes
			// delete them
			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			// Unlock when cleanup is done
			mu.Unlock()
		}
	}()

	// ANY CODE WRITTEN BEFORE THE RETURN ONLY RUNS ONCE

	// Initialize a new rate limiter which allows an average of 2 requests per second,
	// with a maximum of 4 requests in a single ‘burst’.

	// limiter := rate.NewLimiter(2, 4)

	// The function we are returning is a closure, which 'closes over' the limiter
	// variable.
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Only carry out rate limiting if it is enabled.
		if app.config.limiter.enabled {
			// Extract the client's IP address
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}

			// Lock the mutex to pevent this code from being executed concurrently
			mu.Lock()

			// Check to see if the ip address already exists
			// If it doesn't initialize a new rate limiter andd add the IP and limiter to the map
			if _, found := clients[ip]; !found {
				clients[ip] = &client{limiter: rate.NewLimiter(2, 4)}
			}
			clients[ip].lastSeen = time.Now()
			// Call limiter.Allow() to see if the request is permitted, and if it's not,
			// then we call the rateLimitExceededResponse() helper to return a 429 Too Many
			// Requests response (we will create this helper in a minute).
			// if !limiter.Allow() {
			// 	app.rateLimitExceededResponse(w, r)
			// 	return
			// }
			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}

			// Very importantly, unlock the mutex before calling the next handler in the
			// chain. Notice that we DON'T use defer to unlock the mutex, as that would mean
			// that the mutex isn't unlocked until all the handlers downstream of this
			// middleware have also returned.
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

/* MIDDLEWARE FOR USER AUTH */
func (app *application) authenticate(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// Add the "Vary: Authorization" header to the response. This indicates to any
		// caches that the response may vary based on the value of the Authorization header in the request.
		w.Header().Add("Vary", "Authorization")

		// Retrieve the value of the Authorization header from the request. This will
		// return the empty string "" if there is no header found
		authorizationHeader := r.Header.Get("Authorization")

		// If there is no Authorization found. use the contextSetUser() helper
		// to make an Anon user and add it to the request context.
		// Then call the next handler in the chain.
		if authorizationHeader == "" {
			r = app.contextSetUser(r, data.AnonymousUser)
			next.ServeHTTP(w, r)
			return
		}

		// If there is an Auth header it should be in the format "Bearer <token>"
		// Try to split it into it's parts
		// if not in the right format return a 401 to the user
		headerParts := strings.Split(authorizationHeader, " ")
		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Get the token
		token := headerParts[1]

		// Validate tthe token to make sure it is in a sensible format
		v := validator.New()

		// If the token is not valid
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		// Retrive the details of the user associated with the authentication token
		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		// Call the contextSetUser() helper to add the user to the context
		r = app.contextSetUser(r, user)
		next.ServeHTTP(w, r)
	})
}

/*
Notice here that our requireActivatedUser() middleware has a slightly different signature
to the other middleware we’ve built in this book. Instead of accepting and returning a
http.Handler, it accepts and returns a http.HandlerFunc.

This is a small change, but it makes it possible to wrap our /v1/movie** handler functions
directly with this middleware, without needing to make any further conversions
*/
// func (app *application) requiredActivatedUser(next http.HandlerFunc) http.HandlerFunc {

// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

// 		// Use contextGetUser() to retrieve the user from the context
// 		user := app.contextGetUser(r)

// 		// if the user is anonymous, then call the authenticationRequiredREsponse
// 		if user.IsAnonymous() {
// 			app.authenticationRequiredResponse(w, r)
// 			return
// 		}
// 		// if the user is not activated, use the inactiveUserResposne
// 		if !user.Activated {
// 			app.inactiveAccountResponse(w, r)
// 			return
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }

// requireAuthenticatedUser middleware to check that a user is not anonymous
func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Checks that a user is both authenticated and activated.
func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {

	// Rather than returning this http.HandlerFunc we assign it to the variable fn.
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		// Check that a user is activated.

		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
	// Wrap fn with the requireAuthenticatedUser() middleware before returning it.
	return app.requireAuthenticatedUser(fn)
}

// Note that the first parameter for the middleware function is the permission code we require the user to have.
// Also note requirePermission() middleware automatically wraps our existing requireActivatedUser() middleware, which in turn — don’t forget — wraps our requireAuthenticatedUser() middleware.
// So this will do 3 checks.
func (app *application) requirePermission(code string, next http.HandlerFunc) http.HandlerFunc {

	fn := func(w http.ResponseWriter, r *http.Request) {

		// Retrieve the user from the context
		user := app.contextGetUser(r)

		// Get the slice of permissions for the user
		permissions, err := app.models.Permissions.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Check if the slice includes the required permissions
		// if not it's a 401
		if !permissions.Include(code) {
			app.notPermittedResponse(w, r)
			return
		}

		// Here we know they have that permission
		next.ServeHTTP(w, r)

	}

	// Wrap this with requireActivatedUser middlerware before returning it
	return app.requireActivatedUser(fn)
}
