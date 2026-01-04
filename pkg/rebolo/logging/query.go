package logging

import (
	"log"
	"time"
)

// ANSI color codes for terminal output
const (
	ColorReset  = "\033[0m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
)

// LogQuery logs a SQL query in yellow
func LogQuery(query string, args ...interface{}) {
	log.Printf("%s[SQL]%s %s%s%s", ColorYellow, ColorReset, ColorYellow, query, ColorReset)
	if len(args) > 0 {
		log.Printf("%s[SQL Args]%s %v", ColorCyan, ColorReset, args)
	}
}

// LogQueryWithDuration logs a SQL query with execution time
func LogQueryWithDuration(query string, duration time.Duration, args ...interface{}) {
	log.Printf("%s[SQL]%s %s%s%s %s(%v)%s", ColorYellow, ColorReset, ColorYellow, query, ColorReset, ColorCyan, duration, ColorReset)
	if len(args) > 0 {
		log.Printf("%s[SQL Args]%s %v", ColorCyan, ColorReset, args)
	}
}

// LogQueryError logs a SQL query error in red
func LogQueryError(query string, err error, args ...interface{}) {
	log.Printf("%s[SQL ERROR]%s %s%s%s", ColorRed, ColorReset, ColorYellow, query, ColorReset)
	log.Printf("%s[SQL Error]%s %v", ColorRed, ColorReset, err)
	if len(args) > 0 {
		log.Printf("%s[SQL Args]%s %v", ColorCyan, ColorReset, args)
	}
}
