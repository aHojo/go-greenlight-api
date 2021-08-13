package main

import (
	"fmt"
	"net/http"
)

// healthCheckHandler writes a plain text response with info on application status
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Status: available")
	fmt.Fprintln(w, "environment: ", app.config.env)
	fmt.Fprintln(w, "version: ", version)
}