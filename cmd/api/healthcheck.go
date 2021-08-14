package main

import (
	"net/http"
)

// healthCheckHandler writes a plain text response with info on application status
func (app *application) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// Create a fixed-format JSON response
	// js := `{ "status": "available", "environment": %q, "version": %q}`
	js := envelope{ 
		"status":      "available",
		"sys_info": map[string]string{
		"environment": app.config.env,
		"version":     version,
	},
}
	err := app.writeJSON(w, http.StatusOK, js, nil)

	if err != nil {
		app.serverErrorResponse(w,r,err)
		return
	}
}
