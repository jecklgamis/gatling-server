package handler

import (
	"net/http"
)

func BlackholeHandler(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}
