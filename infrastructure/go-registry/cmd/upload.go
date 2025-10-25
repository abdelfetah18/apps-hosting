package cmd

import (
	"go_registry/modulesmanager"
	"go_registry/storage"
	"log"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload [module_version] [module_path]",
	Short: "Upload Go Module in the specified path",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file: %s", err)
		}

		moduleVersion := args[0]
		modulePath := args[1]

		minioStorage := storage.NewMinioStorage(true)
		modulesManager := modulesmanager.NewModulesManager(minioStorage)

		err = modulesManager.UploadModule(moduleVersion, modulePath)
		if err != nil {
			log.Fatalf("Error uploading module: %s", err)
		}

		return nil
	},
}
