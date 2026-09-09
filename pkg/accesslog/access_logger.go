package accesslog

import (
	"log/slog"
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

// AccessLoggerMiddleware logs each request via slog.Default(), i.e. wherever
// the rest of the application's logs go. Use NewAccessLoggerMiddleware to
// send access logs to a separate logger/file instead.
func AccessLoggerMiddleware(next http.Handler) http.Handler {
	return NewAccessLoggerMiddleware(slog.Default())(next)
}

// NewAccessLoggerMiddleware returns access-logging middleware that logs
// through logger, so access logs can be routed to their own file/handler
// independently of the rest of the application's logging.
func NewAccessLoggerMiddleware(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)
			logger.Info("access",
				"host", r.Host,
				"method", r.Method,
				"uri_path", r.RequestURI,
				"protocol", r.Proto,
				"status", lrw.status,
			)
		})
	}
}
