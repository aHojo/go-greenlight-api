package main

import (
	"context"
	"database/sql"
	"flag"
	"os"
	"strings"
	"sync"
	"time"

	// Import the pq driver. It will register itself with the db/sql pacakage.
	"github.com/ahojo/greenlight/internal/data"
	"github.com/ahojo/greenlight/internal/jsonlog"

	"github.com/ahojo/greenlight/internal/mailer"
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
	// Add a new limiter struct containing fields for the requests-per-second and burst
	// values, and a boolean field which we can use to enable/disable rate limiting
	// altogether.
	limiter struct {
		rps     float64
		burst   int
		enabled bool
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		sender   string
	}
	// cors struct to hold our trusted origins
cors struct {
	trustedOrigins []string
}

}

// Holds the dependencies for our http handlers, helpers,
// middleware
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
	mailer mailer.Mailer
	// Sync WaitGroup zero value = waitgroup with a value of 0
	// Don't need to initialize it before use because it's zeroed out.
	wg sync.WaitGroup
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
	// Create command line flags to read the setting values into the config struct.
	// Notice that we use true as the default for the 'enabled' setting?
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")

	// Read the SMTP server configuration settings into the config struct, using the
	// Mailtrap settings as the default values. IMPORTANT: If you're following along,
	// make sure to replace the default values for smtp-username and smtp-password
	// with your own Mailtrap credentials.
	flag.StringVar(&cfg.smtp.host, "smtp-host", "smtp.mailtrap.io", "SMTP host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 2525, "SMTP port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "9dcbdd57b7a1c2", "SMTP username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "8698b43271febb", "SMTP password")
	flag.StringVar(&cfg.smtp.sender, "smtp-sender", "Greenlight <no-reply@greenlight.alexedwards.net>", "SMTP sender")

	// flag.Func() to process the -cors-trusted-origins command line flag
	// Will use strings.Fields() to split the flag value into a slice based on whitespace
	// If no flags are present, then strings.Fields will return an empty slice
	flag.Func("cors-trusted-origins", "Trusted CORS origins (space separated) ", func(val string) error{
		cfg.cors.trustedOrigins = strings.Fields(val)
		return nil
	})
	flag.Parse()

	// Initialize a new logger which writes messages to the standard stream, prefixed
	// with the current date and time
	// logger := log.New(os.Stdout, "", log.Ldate|log.Ltime)
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	db, err := openDB(cfg)
	if err != nil {
		// logger.Fatal(err)
		logger.PrintFatal(err, nil)
	}
	defer db.Close()
	logger.PrintInfo("Database connection pool established", nil)

	// Initialize our application
	app := &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
		mailer: mailer.New(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.sender),
	}
	// // Declare a HTTP server with some sensible timeout settings
	// srv := &http.Server{
	// 	Addr:         fmt.Sprintf(":%d", app.config.port),
	// 	Handler:      app.routes(), // Using the httprouter instance here.
	// 	IdleTimeout:  time.Minute,
	// 	ReadTimeout:  10 * time.Second,
	// 	WriteTimeout: 30 * time.Second,
	// 	// Create a new Go log.Logger instance with log.New()
	// 	// Pass in our custom logger as the first parameter.
	// 	// "" and 0 indicate that the log.Logger instance should not
	// 	// use a prefix or any flags
	// 	ErrorLog: log.New(logger, "", 0),
	// }

	// logger.PrintInfo("Starting Server", map[string]string{
	// 	"addr": srv.Addr,
	// 	"env":  cfg.env,
	// })
	// err = srv.ListenAndServe()

	// Start the server now
	err = app.serve()
	if err != nil {
		logger.PrintFatal(err, nil)
	}

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
