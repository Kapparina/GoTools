package helpers

import (
	"fmt"
	"os"
)

const (
	Success = iota
	ErrReadFile
	ErrWriteFile
	ErrReadWrite
	ErrMoveFile
	ErrStdin
	ErrStdout
	ErrNoInput
	ErrNoFile
	ErrInvalidFileType
	ErrParse
)

// ErrMsg is a custom error type that represents an error and its corresponding Code.
// Err is the error that occurred.
// Code is the Code associated with the error.
// Example usage:
//
//	var processingErr ErrMsg = ErrMsg{nil, Success}
//	defer func() {
//		processingErr.Exit()
//	}()
//
//	// rest of the Code...
type ErrMsg struct {
	Err  error
	Code int
}

// Exit terminates the program with the provided exit Code and prints an error message if there is an error.
// If e.Err is not nil, it prints "An error occurred: <error message>" before exiting.
// It uses defer to ensure that os.Exit is always called, even if an error occurs.
// Example usage:
//
//	var processingErr ErrMsg = ErrMsg{nil, Success}
//	defer func() {
//	  processingErr.Exit()
//	}()
//	...
//	processingErr = processCSV(...)
func (e *ErrMsg) Exit() {
	defer os.Exit(e.Code)
	if e.Err != nil {
		fmt.Printf("An error occured!\nError Code: %d\nDetail: %v\n", e.Code, e.Err)
	}
}
