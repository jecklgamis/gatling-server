package accesslog

import (
	"encoding/json"
	"log"
	"net/http"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.status = code
	lrw.ResponseWriter.WriteHeader(code)
}

func AccessLoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lrw := NewLoggingResponseWriter(w)
		next.ServeHTTP(lrw, r)
		accessLog := map[string]interface{}{"host": r.Host,
			"method": r.Method, "uri_path": r.RequestURI, "protocol": r.Proto,
			"status": lrw.status,
		}
		bytes, _ := json.Marshal(accessLog)
		log.Println(string(bytes))
	})
}
