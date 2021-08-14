package data

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Error if we can't parse or convert to the JSON string successfully
var ErrInvalidRuntimeFormat = errors.New("invalid runtime format")

type Runtime int32

// Implement our MarshalJSON() method.
// This satisfies the Marshaler interface and will be called when json.Marshal is called
func (r Runtime) MarshalJSON() ([]byte, error) {

	// Generate a string containing the movie runtime required format
	jsonValue := fmt.Sprintf("%d mins", r)

	// use the strconv.Quote() to wrap it in double quotes.
	// this will fail to marshal as json without this
	quotedJSONValue := strconv.Quote(jsonValue)

	// convert to []byte and return
	return []byte(quotedJSONValue), nil
}

// Implemenrt the UnmarshalJSON() method for the json.Unmarshaler interface
// IMPORTANT: use UnmarshalJSON() needs to modify the 
// receiver (our Runtime type), we must use a pointer receiver for this to work
// correctly. 
func (r *Runtime) UnmarshalJSON(jsonValue []byte) error {

	// We expect that the incoming JSON value will be a string in the format
	// "<runtime> mins", and the first thing we need to do is remove the surrounding // double-quotes from this string. If we can't unquote it, then we return the
	// ErrInvalidRuntimeFormat error.
	unquotedJSONValue, err := strconv.Unquote(string(jsonValue))
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Split the string to isolate the part containing the number
	parts := strings.Split(unquotedJSONValue, " ")

	// Sanity check  the parts of the string to make sure it was in the expected format.
	if len(parts) != 2 || parts[1] != "mins" {
		return ErrInvalidRuntimeFormat
	}

	// parse the string into an int32 again
	i, err := strconv.ParseInt(parts[0], 10, 32)
	if err != nil {
		return ErrInvalidRuntimeFormat
	}

	// Convert the int32 to a Runtime type and assing this to the reciever.
	*r = Runtime(i)
	return nil
}
