package handler

import (
	"crypto/subtle"
	"fmt"
	"net/http"
)

// RequireAuth wraps next with the same bearer-token check and per-client
// rate limiting used by the rest of this package, for handlers that aren't
// otherwise gated by a dedicated handler type.
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

// RequireBasicAuth wraps next with an HTTP Basic Auth check and the same
// per-client rate limiting as RequireAuth. Basic Auth (rather than the
// bearer token used elsewhere) is used for the /workspace/ and /uploads/
// file listings specifically because they're meant to be browsed directly
// in a browser: the browser's native login prompt and credential caching
// let a user click through links without attaching a custom header by hand.
func RequireBasicAuth(next http.Handler, username, password string) http.Handler {
	limiter := newAuthLimiter()
	const realm = "gatling-server"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		clientKey := clientIP(r)
		if limiter.blocked(clientKey) {
			tooManyRequestsWithError(w, fmt.Errorf("too many failed authentication attempts"))
			return
		}
		reqUser, reqPass, ok := r.BasicAuth()
		authorized := ok && username != "" && password != "" &&
			subtle.ConstantTimeCompare([]byte(reqUser), []byte(username)) == 1 &&
			subtle.ConstantTimeCompare([]byte(reqPass), []byte(password)) == 1
		if !authorized {
			limiter.recordFailure(clientKey)
			w.Header().Set("WWW-Authenticate", fmt.Sprintf("Basic realm=%q", realm))
			unauthorizedWithError(w, fmt.Errorf("missing or invalid credentials"))
			return
		}
		next.ServeHTTP(w, r)
	})
}
