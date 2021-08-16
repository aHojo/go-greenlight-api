package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	// Import the pq driver. It will register itself with the db/sql pacakage.
	"github.com/ahojo/greenlight/internal/data"
	_ "github.com/lib/pq"
)

// The version of our api
const version = "1.0.0"

// All the config settings for our application
type config struct {
	port int
	env  string
	db   struct {
		dsn          string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime  string
	}
}

// Holds the dependencies for our http handlers, helpers,
// middleware
type application struct {
	config config
	logger *log.Logger
	models data.Models
}

func main() {
	var cfg config

	// Get the port and env from the commandline
	flag.IntVar(&cfg.port, "port", 4000, "The port our api will listen on")
	flag.StringVar(&cfg.env, "env", "dev", "The environment our code was built with.")

	// DB CLI flags
	flag.StringVar(&cfg.db.dsn, "db-dsn", os.Getenv("GREENLIGHT_DB_DSN"), "PostgresSQL DSN")
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgresSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idleconns", 25, "PostgresSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgresSQL max connection idle time")
	flag.Parse()

	// Initialize a new logger which writes messages to the standard stream, prefixed
	// with the current date and time
	logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)

	db, err := openDB(cfg)
	if err != nil {
		logger.Fatal(err)
	}
	defer db.Close()

	// Initialize our application
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}
	// Declare a HTTP server with some sensible timeout settings
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(), // Using the httprouter instance here.
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	app.logger.Printf("Starting %s server on %s", app.config.env, srv.Addr)
	err = srv.ListenAndServe()
	app.logger.Fatal(err)
}

func openDB(cfg config) (*sql.DB, error) {
	// create an open db connection pool.
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err != nil {
		return nil, err
	}

	// Set the maximum number of open (in use + idle) connections in the pool.
	// <= 0 means unlimited
	db.SetMaxOpenConns(cfg.db.maxOpenConns)

	// Set the maximum number of idle connections
	// <= 0 means unlimited
	db.SetMaxIdleConns(cfg.db.maxIdleConns)

	// use the time.ParseDuration() funciton to convert the idle timeout duration string to a time.Duration
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(duration)

	// Create a context with a 5-second timeout deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// use PingContext to establish a new connection to the database, passing the contect we created above.
	// if in 5 seconds no connection is established, error.
	err = db.PingContext(ctx)
	if err != nil {
		return nil, err
	}

	// return the pool
	return db, nil
}
