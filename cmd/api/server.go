package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*
Signal   Description                         Keyboard shortcut  Catchable
SIGINT   Interrupt from keyboard              Ctrl+C 						 Yes
SIGQUIT Quit from keyboard 					          Ctrl+\						 Yes
SIGKILL Kill process (terminate immediately)    -			           No
SIGTERM Terminate process in orderly manner 		-			           Yes
*/

func (app *application) serve() error {

	// Declare a HTTPSERVER using the same settings as in our main() function
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(), // Using the httprouter instance here.
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		// Create a new Go log.Logger instance with log.New()
		// Pass in our custom logger as the first parameter.
		// "" and 0 indicate that the log.Logger instance should not
		// use a prefix or any flags
		ErrorLog: log.New(app.logger, "", 0),
	}

	// Create a shutdownError channel. Will recieve any errors returnd by the graceful shutdown
	shutdownError := make(chan error)

	// Background go routine that will catch our interrupts.
	go func() {
		// create a quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)

		// signal.Notify() to listen for incoming SIGINT and SIGTERM signals
		// Relay them to the quit channel.
		// Other signals will not be caught
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// Read the singal from the channel. Blocks until channel is received
		s := <-quit

		// Log a message that the signal has been caught
		// The string() method to get the signal name and include it in the log entry
		app.logger.PrintInfo("shutting down the server", map[string]string{
			"signal": s.String(),
		})

		// Create a context with a 5-second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// // Shutdown() passing in the context
		// // returns nil if the graceful shutdown was successful
		// // returns an error if a problem closing the listeners, or
		// // we hit the context deadline.=
		// shutdownError <- srv.Shutdown(ctx)
		
		// Since we are using Waitgroups now, only send the shutdown on an error
		err := srv.Shutdown(ctx)
		if err != nil {
		  shutdownError <- err
		}

		// Tell the logger that we are waiting for background tasks to complete
		app.logger.PrintInfo("Completing background tasks", map[string]string{
			"addr": srv.Addr,
		})


		// Call the Wait() to block until Waitgroup counter is 0
		// Block until all goroutines have finished. 
		// Return nil on the shutdownError channel, to indicate the shutdown completed with no issues
		app.wg.Wait()
		shutdownError <- nil

	}()

	// Calling Shutdown() will cause ListenAndServe to immediately return http.ErrServerClosed
	// If we see this error the graceful shutdown has started
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// We wait to recieve the return value from shutdown on the shutdown channel.
	// If there is an error, there was a problem
	err = <-shutdownError
	if err != nil {
		return err
	}

	// Likewise log a "starting server" message
	app.logger.PrintInfo("Stopped server", map[string]string{
		"addr": srv.Addr,
	})

	return nil
}
