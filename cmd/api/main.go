package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// The version of our api
const version  = "1.0.0"

// All the config settings for our application
type config struct {
	port int
	env string
}

// Holds the dependencies for our http handlers, helpers,
// middleware
type application struct {
	config config
	logger *log.Logger
}
func main() {
	var cfg config

	// Get the port and env from the commandline
	flag.IntVar(&cfg.port, "port",4000, "The port our api will listen on")
	flag.StringVar(&cfg.env, "env", "dev", "The environment our code was built with.")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard stream, prefixed
	// with the current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// Initialize our application
	app := &application{
		config: cfg,
		logger: logger,
	}



	// Declare a HTTP server with some sensible timeout settings
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(), // Using the httprouter instance here.
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.logger.Printf("Starting %s server on %s", app.config.env, srv.Addr)
	err := srv.ListenAndServe()
	app.logger.Fatal(err)
}