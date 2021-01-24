package handler

import (
	"net/http"
)

func RootHandler(w http.ResponseWriter, _ *http.Request) {
	okWithJson(w, map[string]interface{}{"name": "gatling-server", "message": "Relax, perf test it."})
}
