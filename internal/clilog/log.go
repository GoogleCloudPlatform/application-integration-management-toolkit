package clilog

import (
	"io"
	"log"
	"os"
)

// log levels, default is error
var (
	Debug        *log.Logger
	Info         *log.Logger
	Warning      *log.Logger
	Error        *log.Logger
	HttpResponse *log.Logger
	HttpError    *log.Logger
)

// Init function initializes the logger objects
func Init(debug bool, print bool, noOutput bool, suppressWarnings bool) {
	var debugHandle = io.Discard
	var infoHandle = io.Discard
	var warningHandle, errorHandle, responseHandle io.Writer

	if debug {
		debugHandle = os.Stdout
	}

	if print {
		infoHandle = os.Stdout
	}

	if noOutput {
		responseHandle = io.Discard
		infoHandle = io.Discard
		errorHandle = io.Discard
		warningHandle = io.Discard
	} else {
		responseHandle = os.Stdout
		warningHandle = os.Stdout
		errorHandle = os.Stderr
	}

	if suppressWarnings {
		warningHandle = io.Discard
	}

	Debug = log.New(debugHandle,
		"DEBUG: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"", 0)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	HttpResponse = log.New(responseHandle,
		"", 0)

	HttpError = log.New(errorHandle,
		"", 0)
}
