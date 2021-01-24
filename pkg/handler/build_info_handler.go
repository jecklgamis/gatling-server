package handler

import (
	"github.com/jecklgamis/gatling-server/pkg/version"
	"net/http"
)

func BuildInfoHandler(w http.ResponseWriter, _ *http.Request) {
	okWithJson(w, map[string]interface{}{"name": "gatling-server",
		"version": version.BuildVersion, "branch": version.BuildBranch})
}
