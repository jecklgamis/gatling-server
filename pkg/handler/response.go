package handler

import (
	"encoding/json"
	"log"
	"net/http"
)

func methodNotAllowed(w http.ResponseWriter) {
	w.WriteHeader(http.StatusMethodNotAllowed)
}

func badRequest(w http.ResponseWriter) {
	w.WriteHeader(http.StatusBadRequest)
}

func badRequestWithJson(w http.ResponseWriter, entity map[string]interface{}) {
	w.WriteHeader(http.StatusBadRequest)
	err := json.NewEncoder(w).Encode(entity)
	if err != nil {
		log.Println(err)
	}
}

func badRequestWithError(w http.ResponseWriter, error error) {
	entity := map[string]interface{}{
		"ok":    false,
		"error": error}
	badRequestWithJson(w, entity)
}

func notFound(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
}

func internalServerError(w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
}

func okWithJson(w http.ResponseWriter, entity interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(entity)
	if err != nil {
		log.Println(err)
	}
}

func okWithEntity(w http.ResponseWriter, contentType string, data []byte) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func ok(w http.ResponseWriter) {
	w.WriteHeader(http.StatusOK)
}
