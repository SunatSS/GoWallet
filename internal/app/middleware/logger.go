package middleware

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
)

// Logger is a middleware fuction that logs path and method of request
func Logger(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Start: %s %s", r.Method, r.URL.Path)
		handle.ServeHTTP(w, r)
		log.Printf("FINISH: %s %s", r.Method, r.URL.Path)
	})
}

var ErrNoFile = errors.New("no log file")

var loggerContextKey = &loggerKey{"logger context"}

type loggerKey struct {
	name string
}

func (l *loggerKey) String() string {
	return l.name
}

type Loggers struct {
	InfoLogger  *log.Logger
	ErrorLogger *log.Logger
}

// LoggersFuncs is a middleware function that logs info to file
func LoggersFuncs(handle http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f, err := os.OpenFile("../log.log", os.O_APPEND|os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		f.Write([]byte("\r\n"))

		ctx := context.WithValue(r.Context(), loggerContextKey, f)
		r = r.WithContext(ctx)

		handle.ServeHTTP(w, r)
	})
}

func GetLoggers(ctx context.Context) (*Loggers, error) {
	value, ok := ctx.Value(loggerContextKey).(*os.File)
	if !ok {
		return nil, ErrNoFile
	}

	var loggers = Loggers{
		log.New(value, "INFO\t", log.Ldate|log.Ltime),
		log.New(value, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile),
	}
	
	return &loggers, nil
}
