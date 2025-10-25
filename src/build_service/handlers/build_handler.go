package handlers

import (
	"net/http"
	"os"
	"path/filepath"

	"apps-hosting.com/messaging"

	"github.com/gorilla/mux"
)

type BuildHandler struct{}

func NewBuildHandler() BuildHandler {
	return BuildHandler{}
}

func (handler *BuildHandler) HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	messaging.WriteSuccess(w, "OK", nil)
}

func (handler *BuildHandler) DownloadSourceCode(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	repoId := params["repo_id"]

	fileName := repoId
	downloadPath := "/shared/repos/"

	filePath := filepath.Join(downloadPath, fileName)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		messaging.WriteError(w, http.StatusNotFound, "File not found")
		return
	}

	// Set headers and serve the file
	w.Header().Set("Content-Disposition", "attachment; filename="+fileName)
	w.Header().Set("Content-Type", "application/octet-stream")

	http.ServeFile(w, r, filePath)
}
