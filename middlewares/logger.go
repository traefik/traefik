package middlewares

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
)

// Logger is a middleware handler that logs the request as it goes in and the response as it goes out.
type Logger struct {
	file *os.File
}

// NewLogger returns a new Logger instance.
func NewLogger(file string) *Logger {
	if len(file) > 0 {
		fi, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			log.Fatal("Error opening file", err)
		}
		return &Logger{fi}
	}
	return &Logger{nil}
}

func (l *Logger) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if l.file == nil {
		next(rw, r)
	} else {
		handlers.CombinedLoggingHandler(l.file, next).ServeHTTP(rw, r)
	}
}

// Close closes the logger (i.e. the file).
func (l *Logger) Close() {
	if l.file != nil {
		l.file.Close()
	}
}
