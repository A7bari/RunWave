package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Error struct {
	orig error
	msg  string
	code ErrorCode
}

type ErrorCode uint

const (
	ErrorCodeTimeout ErrorCode = iota
	ErrorCodePodNotFound
	ErrorCodePodUpdate
	ErrorCodeUnsupportLanguage
	ErrorCodeExecutionErr
)

// WrapErrorf returns a wrapped error.
func WrapErrorf(orig error, code ErrorCode, format string, a ...interface{}) error {
	return &Error{
		code: code,
		orig: orig,
		msg:  fmt.Sprintf(format, a...),
	}
}

// NewErrorf instantiates a new error.
func NewErrorf(code ErrorCode, format string, a ...interface{}) error {
	return WrapErrorf(nil, code, format, a...)
}

// Error returns the message, when wrapping errors the wrapped error is returned.
func (e *Error) Error() string {
	if e.orig != nil {
		return fmt.Sprintf("%s: %v", e.msg, e.orig)
	}

	return e.msg
}

// Unwrap returns the wrapped error, if any.
func (e *Error) Unwrap() error {
	return e.orig
}

// Code returns the code representing this error.
func (e *Error) Code() ErrorCode {
	return e.code
}

// general errors
var (
	ErrNoPodsAvailable = NewErrorf(ErrorCodePodNotFound, "no standby pods available")
)

// respondWithJSONError is a helper function to send a structured JSON error response
func RespondWithJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Message: message,
		Code:    statusCode,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to send error response: %v", err)
	}
}
