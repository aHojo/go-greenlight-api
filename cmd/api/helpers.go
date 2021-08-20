package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/ahojo/greenlight/internal/validator"
	"github.com/julienschmidt/httprouter"
)

type envelope map[string]interface{}

// readIDParam - gets the ID URL parameter from the current context
func (app *application) readIDParam(r *http.Request) (int64, error) {
	// When httprouter parses a request, interpolated parameters will be stored
	// in the request context. Use ParamsFromContext() function to
	// get the slice containing them.
	params := httprouter.ParamsFromContext(r.Context())

	// Use the ByName() method to get the value of the id parameter from the slice
	// ByName() always returns a string, se we will convert it to a base 10 int
	// if the id is invalid return 404
	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}
	return id, nil
}

// WriteJSON - writes a JSON response to the response writer
// Takes HTTP status code, data to encode to JSON, and a header map for additional header
func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// Encode the data to JSON
	// js, err := json.Marshal(data)

	// We will use MarshalIndent to make the format better
	js, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// No more errors can occur so add the headers
	// No errors occur if the map is nil
	for key, value := range headers {
		w.Header()[key] = value
	}

	// Add Content-Type header
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	// limit the size to 1MB
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	// Initialize the json.Decoder. call the DisallowedUnknownFields() method on it before decoding.
	// If there is a field that is not in our map(dst) there will be an error
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	// Decode the body into the destination
	err := dec.Decode(dst)

	if err != nil {
		// if there is a syntax error
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		// Use the error.As() function to check whether the error has the type *json.SyntaxError
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)
		// If Decode() returns an io.ErrUnexpectedEOR
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")
		case errors.As(err, &unmarshalTypeError):
			return fmt.Errorf("body contains incorrect JSON type field %q", unmarshalTypeError.Field)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")
		case errors.As(err, &invalidUnmarshalError):
			panic(err)
		// if the JSON contains a field which can't be mapped to the target destination
		//Decode() will now return an error message in the format "json: unknown // field "<name>"".
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		// If the body exceeds maxBytes decode will fail "http: request body too large"
		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must be not larger than %d bytes", maxBytes)
		default:
			return err
		}
	}

	// Call Decode() again, using a pointer to an empty anon struct as the destination.
	// If the r.Body only contained a single JSON value, this will return an io.EOF
	// if anything else we don't process and send an error.
	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// readString() returns a string value from the query string, or
// a default value if no matching key could be found.
func (app *application) readString(qs url.Values, key string, defaultValue string) string {

	// Extract the value for a given key from the query string
	// Default value if no key exists
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	return s
}

// readCSV() reads a string value from the query string
// splits it into a slice on the comma.
// returns a default value if it can not be found
func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {

	// Extract the value from the query string
	csv := qs.Get(key)
	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

// readInt() helper reads a string value from the query string and converts it to an
// integer before returning. If no matching key could be found it returns the provided
// default value. If the value couldn't be converted to an integer, then we record an
// error message in the provided Validator instance.
func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)
	if s == "" {
		return defaultValue
	}

	// Try to convert to int
	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}
	
	return i
}


// The background() helper accepts an arbitray function as a param
func (app *application) background(fn func()) {
  // Increment the waitgroup by 1 since we are starting a goroutine	
	app.wg.Add(1)

	// launch the backgound go routine	
	go func (){

		// Decrement the counter before the goroutine returns
		defer app.wg.Done()
		// panic recovery
		defer func() {
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		// execute whatever function we passed in here.
		fn()
	}() 
}
