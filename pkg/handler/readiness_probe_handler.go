package handler

import (
	"net/http"
)

func ReadinessProbeHandler(w http.ResponseWriter, _ *http.Request) {
	okWithJson(w, map[string]interface{}{"message": "I'm ready!"})
}
