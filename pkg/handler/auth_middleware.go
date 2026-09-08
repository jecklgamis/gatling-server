package handler

import (
	"fmt"
	"net/http"
)

// RequireAuth wraps next with the same bearer-token check and per-client
// rate limiting used by the rest of this package, for handlers (like the
// /workspace/ static file server) that aren't otherwise gated by a
// dedicated handler type.
func RequireAuth(next http.Handler, apiToken string) http.Handler {
	limiter := newAuthLimiter()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientKey := clientIP(r)
		if limiter.blocked(clientKey) {
			tooManyRequestsWithError(w, fmt.Errorf("too many failed authentication attempts"))
			return
		}
		if !isAuthorized(r, apiToken) {
			limiter.recordFailure(clientKey)
			unauthorizedWithError(w, fmt.Errorf("missing or invalid API token"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
