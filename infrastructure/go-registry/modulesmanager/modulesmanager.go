package modulesmanager

import (
	"bufio"
	"encoding/json"
	"fmt"
	"go_registry/storage"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/module"
	"golang.org/x/mod/zip"
)

type ModulesManager struct {
	storage storage.Storage
}

func NewModulesManager(storage storage.Storage) *ModulesManager {
	return &ModulesManager{storage: storage}
}

func (m *ModulesManager) UploadModule(moduleVersion string, modulePath string) error {
	// Ensure module path exists
	if _, err := os.Stat(modulePath); os.IsNotExist(err) {
		return fmt.Errorf("module path does not exist: %s", modulePath)
	}

	// Parse go.mod to get the module path
	modFile := filepath.Join(modulePath, "go.mod")
	modPath, err := readModulePath(modFile)
	if err != nil {
		return fmt.Errorf("failed to read module path from go.mod: %w", err)
	}

	// Check If a version already exists
	hasFile := m.storage.HasFile(filepath.Join(modPath, fmt.Sprintf("%s.info", moduleVersion)))
	if hasFile {
		return fmt.Errorf("module with that version already exist.")
	}

	// Create unique output directory
	outputDir, err := os.MkdirTemp("", "gomod_upload_*")
	if err != nil {
		return fmt.Errorf("failed to create temp output directory: %w", err)
	}

	// --- 1. Create $version.info file ---
	info := struct {
		Version string    `json:"Version"`
		Time    time.Time `json:"Time"`
	}{
		Version: moduleVersion,
		Time:    time.Now(),
	}

	infoData, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal info JSON: %w", err)
	}

	infoFile := filepath.Join(outputDir, fmt.Sprintf("%s.info", moduleVersion))
	if err := os.WriteFile(infoFile, infoData, 0644); err != nil {
		return fmt.Errorf("failed to write info file: %w", err)
	}

	// --- 2. Create $version.mod file ---
	modFileDst := filepath.Join(outputDir, fmt.Sprintf("%s.mod", moduleVersion))
	if err := copyFile(modFile, modFileDst); err != nil {
		return fmt.Errorf("failed to copy go.mod file: %w", err)
	}

	// --- 3. Create $version.zip file ---
	zipFile := filepath.Join(outputDir, fmt.Sprintf("%s.zip", moduleVersion))
	outFile, err := os.Create(zipFile)
	if err != nil {
		return fmt.Errorf("failed to create zip file: %w", err)
	}
	defer outFile.Close()

	modVer := module.Version{
		Path:    modPath,
		Version: moduleVersion,
	}

	if err := zip.CreateFromDir(outFile, modVer, modulePath); err != nil {
		return fmt.Errorf("failed to create module zip: %w", err)
	}

	fmt.Println("✅ Module upload package created successfully:")
	fmt.Println("  Module Path:", modPath)
	fmt.Println("  ", infoFile)
	fmt.Println("  ", modFileDst)
	fmt.Println("  ", zipFile)

	err = m.storage.PutFile(filepath.Join(modPath, filepath.Base(infoFile)), infoFile)
	if err != nil {
		panic(err)
	}

	err = m.storage.PutFile(filepath.Join(modPath, filepath.Base(modFileDst)), modFile)
	if err != nil {
		panic(err)
	}

	err = m.storage.PutFile(filepath.Join(modPath, filepath.Base(zipFile)), zipFile)
	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Module uploaded successfully")

	return err
}

// copyFile copies a file from src to dst.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// readModulePath parses the module path from a go.mod file.
func readModulePath(modFile string) (string, error) {
	f, err := os.Open(modFile)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			path := strings.TrimSpace(strings.TrimPrefix(line, "module "))
			path = strings.Trim(path, `"`) // remove quotes if any
			if path == "" {
				return "", fmt.Errorf("module path is empty in go.mod")
			}
			return path, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", err
	}

	return "", fmt.Errorf("no module path found in go.mod")
}
