package hive

import (
	"fmt"
	"strings"
)

// Error codes from package
const (
	ErrInternal            = "INTERNAL"
	ErrInvalidJSON         = "INVALID_JSON"
	ErrInvalidLoginRespose = "INVALID_LOGIN_RESPONSE"
	ErrInvalidNodeType     = "INVALID_NODE_TYPE"
	ErrInvalidNodeJSON     = "INVALID_NODE_JSON"
	ErrInvalidDataType     = "INVALID_DATA_TYPE"
	ErrNodeNotFound        = "NODE_NOT_FOUND"
	ErrInvalidUpdate       = "INVALID_UPDATE"
)

// Error codes from Hive API
const (
	ErrMissingParameter   = "MISSING_PARAMETER"
	ErrInvalidCredentials = "USERNAME_PASSWORD_ERROR"
	ErrNotAuthorized      = "NOT_AUTHORIZED"
)

// Error defines a standard application error.
type Error struct {
	// Machine-readable error code.
	Code string

	// Human-readable message.
	Message string

	// Logical operation and nested error.
	Op  string
	Err error
}

// Error returns the string representation of the error message.
func (e *Error) Error() string {
	var buf strings.Builder

	// Print the current operation in our stack, if any.
	if e.Op != "" {
		fmt.Fprintf(&buf, "%s: ", e.Op)
	}

	// If wrapping an error, print its Error() message.
	// Otherwise print the error code & message.
	if e.Err != nil {
		buf.WriteString(e.Err.Error())
	} else {
		if e.Code != "" {
			fmt.Fprintf(&buf, "<%s> ", e.Code)
		}

		buf.WriteString(e.Message)
	}

	return buf.String()
}

// ErrorCode returns the code of the root error, if available. Otherwise returns ErrInternal.
func ErrorCode(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Code != "" {
		return e.Code
	} else if ok && e.Err != nil {
		return ErrorCode(e.Err)
	}

	return ErrInternal
}

// ErrorMessage returns the human-readable message of the error, if available.
// Otherwise returns a generic error message.
func ErrorMessage(err error) string {
	if err == nil {
		return ""
	} else if e, ok := err.(*Error); ok && e.Message != "" {
		return e.Message
	} else if ok && e.Err != nil {
		return ErrorMessage(e.Err)
	}

	return "internal error has occurred"
}
