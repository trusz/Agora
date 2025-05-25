package log

import (
	"encoding/json"
	"log"
	"os"
)

var (
	// Debug _
	Debug *log.Logger = log.New(
		os.Stdout,
		"DEBUG: ",
		log.Lshortfile,
	)
	// Trace _
	Trace *log.Logger = log.New(
		os.Stdout,
		"TRACE : ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
	// Info _
	Info *log.Logger = log.New(
		os.Stdout,
		"INFO: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
	// Warning _
	Warning *log.Logger = log.New(
		os.Stdout,
		"⚠️ WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
	// Error _
	Error *log.Logger = log.New(
		os.Stderr,
		"🔥  ERROR: ",
		log.Ldate|log.Ltime|log.Llongfile,
	)
	// Test _
	Test *log.Logger = log.New(
		os.Stderr,
		"TEST: ",
		log.Ldate|log.Ltime|log.Lshortfile,
	)
	// Pretty _
	Pretty = func(template string, objs ...interface{}) {
		prettyString, err := json.MarshalIndent(objs, "", "  ")
		if err != nil {
			Error.Printf("error: %s", err)
		}

		Debug.Printf("%s", prettyString)
	}
)
