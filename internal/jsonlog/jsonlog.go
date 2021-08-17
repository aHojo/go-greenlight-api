package jsonlog

import (
	"encoding/json"
	"io"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

// Represents the severity level for a log entry
type Level int8

// Constants which represent a specific security level.
const (
	LevelInfo  Level = iota // Has the value 0
	LevelError              // Has the value 1
	LevelFatal              // Has the value 2
	LevelOff                // Has the value 3
)

// String - return a human-friendly string for the severity level
func (l Level) String() string {
	switch l {
	case LevelInfo:
		return "INFO"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return ""
	}
}

// Logger - Holds the output destination that the log entries
// will be written to. The minimum severity level that log entries will be written for,
// plus a mutex for coordination writes.
type Logger struct {
	out      io.Writer
	minLevel Level
	mu       sync.Mutex
}

// New - return a new logger which writes log entries at or above minimum security level
func New(out io.Writer, minLevel Level) *Logger {
	return &Logger{
		out: out,
		minLevel: minLevel,
	}
}

// Declare some helper methods for writing log entries at the different levels. Notice
// that these all accept a map as the second parameter which can contain any arbitrary
// 'properties' that you want to appear in the log entry

func (l *Logger) PrintInfo(message string, properties map[string]string) {
	l.print(LevelInfo, message, properties)
}
func (l *Logger) PrintError(err error, properties map[string]string) {
	l.print(LevelError, err.Error(), properties)
}
func (l *Logger) PrintFatal(err error, properties map[string]string) {
	l.print(LevelFatal, err.Error(), properties)
	os.Exit(1) // For entries at Fatal level, terminate the application
}

// Print is an internal method for writing the log entry
func (l *Logger) print(level Level, message string, properties map[string]string) (int, error) {
	// If the severity level of the log entry is below the minimum severity for the logger
	// return with no further action
	if level < l.minLevel {
		return 0, nil
	}

	// Anon struct to hold the data for the log entry
	aux := struct {
		Level      string            `json:"level"`
		Time       string            `json:"time"`
		Message    string            `json:"message"`
		Properties map[string]string `json:"properties,omitempty"`
		Trace      string            `json:"trace,omitempty"`
	}{
		Level:      level.String(),
		Time:       time.Now().UTC().Format(time.RFC3339),
		Message: message,
		Properties: properties,
	}

	// Include a stack trace for entries at the Error and FATAL levels.
	if level >= LevelError {
		aux.Trace = string(debug.Stack())
	}

	// Declare a line variable  for holding the acutal log entry text
	var line []byte

	// Marshal the anonymous struct to JSON and store it in the line variable. If there
	// was a problem creating the JSON, set the contents of the log entry to be that
	// plain-text error message instead.
	line, err := json.Marshal(aux)
	if err != nil {
		line = []byte(LevelError.String() + ": unable to marshal log message:" + err.Error())
	}

	// Lock the mutex
	// Two writes to the output dest can not happen
	l.mu.Lock()
	defer l.mu.Unlock()

	return l.out.Write(append(line, '\n'))

}

// Write method on our Logger type so that it satisfies the 
// io.Writer interface. This writes a log entry at the ERROR level with no additional 
// properties.
func (l *Logger) Write(message []byte) (n int, err error) {
	return l.print(LevelError, string(message), nil)
}
