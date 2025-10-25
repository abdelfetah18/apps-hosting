package handler

import (
	"encoding/json"
	"fmt"
	"go_registry/storage"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

type UploadBody struct {
	Project    string
	Repository string
	ModulePath string
	Version    string
	Source     string
}

type Handler struct {
	storage storage.Storage
}

func NewHandler(storage storage.Storage) *Handler {
	return &Handler{storage: storage}
}

func (handler *Handler) ListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	base := vars["base"]
	module := vars["module"]

	fmt.Printf("ListHandler: base=%s, module=%s\n", base, module)

	files := handler.storage.ListFiles(filepath.Join(base, module))
	versions := []string{}
	for _, file := range files {
		if strings.Contains(file, ".info") {
			version := strings.Split(strings.Split(file, ".info")[0], module+"/")[1]
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 Not Found: object not found"))
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(strings.Join(versions, "\n")))
}

func (handler *Handler) InfoHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	base := vars["base"]
	module := vars["module"]
	version := vars["version"]

	fmt.Printf("InfoHandler: base=%s, module=%s, version=%s\n", base, module, version)

	objectName := filepath.Join(base, module, fmt.Sprintf("%s.info", version))

	if !handler.storage.HasFile(objectName) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 Not Found: object not found"))
		return
	}

	object, err := handler.storage.GetFile(objectName)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 Internal Server Error"))
		return
	}

	type Info struct {
		Version string    `json:"version"`
		Time    time.Time `json:"time"`
	}

	var info Info
	if err := json.NewDecoder(object).Decode(&info); err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 Internal Server Error: failed to decode info file"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(info); err != nil {
		fmt.Printf("Error encoding response: %v\n", err)
	}
}

func (handler *Handler) ModHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	base := vars["base"]
	module := vars["module"]
	version := vars["version"]

	fmt.Printf("ModHandler: base=%s, module=%s, version=%s\n", base, module, version)

	objectName := filepath.Join(base, module, fmt.Sprintf("%s.mod", version))

	if !handler.storage.HasFile(objectName) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 Not Found: object not found"))
		return
	}

	object, err := handler.storage.GetFile(objectName)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	// Copy object stream to response writer
	if _, err := io.Copy(w, object); err != nil {
		fmt.Printf("Failed to write object to response: %v\n", err)
	}
}

func (handler *Handler) ZipHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	base := vars["base"]
	module := vars["module"]
	version := vars["version"]

	fmt.Printf("ZipHandler: base=%s, module=%s, version=%s\n", base, module, version)

	objectName := filepath.Join(base, module, fmt.Sprintf("%s.zip", version))

	if !handler.storage.HasFile(objectName) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("404 Not Found: object not found"))
		return
	}

	object, err := handler.storage.GetFile(objectName)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("500 Internal Server Error"))
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.WriteHeader(http.StatusOK)

	// Copy object stream to response writer
	if _, err := io.Copy(w, object); err != nil {
		fmt.Printf("Failed to write object to response: %v\n", err)
	}
}

func (handler *Handler) LatestHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	base := vars["base"]
	module := vars["module"]

	fmt.Printf("LatestHandler: base=%s, module=%s\n", base, module)

	files := handler.storage.ListFiles(filepath.Join(base, module))
	versions := []string{}
	for _, file := range files {
		if strings.Contains(file, ".info") {
			version := strings.Split(strings.Split(file, ".info")[0], module+"/")[1]
			versions = append(versions, version)
		}
	}

	if len(versions) == 0 {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("404 Not Found"))
		return
	}

	latestVersion := LatestVersion(versions)

	objectName := filepath.Join(base, module, fmt.Sprintf("%s.info", latestVersion))

	object, err := handler.storage.GetFile(objectName)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("404 Not Found"))
		return
	}

	type Info struct {
		Version string    `json:"version"`
		Time    time.Time `json:"time"`
	}

	info := Info{}
	json.NewDecoder(object).Decode(&info)

	response := Info{
		Version: info.Version,
		Time:    info.Time,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (handler *Handler) WriteResponse(w http.ResponseWriter) {
	type Response struct {
		Message string
	}

	response := Response{
		Message: "Hello World",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// LatestVersion takes a list of version strings (e.g., ["v0.0.1", "v1.0.0", "v2.0.1"])
// and returns the latest one based on semantic version comparison.
func LatestVersion(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	// Helper to parse version parts into integers
	parseVersion := func(v string) ([]int, error) {
		v = strings.TrimPrefix(v, "v")
		parts := strings.Split(v, ".")
		nums := make([]int, len(parts))
		for i, p := range parts {
			n, err := strconv.Atoi(p)
			if err != nil {
				return nil, err
			}
			nums[i] = n
		}
		return nums, nil
	}

	sort.Slice(versions, func(i, j int) bool {
		vi, _ := parseVersion(versions[i])
		vj, _ := parseVersion(versions[j])

		for k := 0; k < len(vi) && k < len(vj); k++ {
			if vi[k] != vj[k] {
				return vi[k] < vj[k]
			}
		}
		return len(vi) < len(vj)
	})

	return versions[len(versions)-1]
}
