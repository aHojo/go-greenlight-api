package main

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

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
				if time.Since(client.lastSeen) > 3 * time.Minute {
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
			clients[ip] = &client{ limiter: rate.NewLimiter(2, 4)}
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

		next.ServeHTTP(w, r)
	})
}
