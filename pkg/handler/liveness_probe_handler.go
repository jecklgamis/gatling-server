package handler

import (
	"net/http"
)

func LivenessProbeHandler(w http.ResponseWriter, _ *http.Request) {
	okWithJson(w, map[string]interface{}{"message": "I'm alive!"})
}
