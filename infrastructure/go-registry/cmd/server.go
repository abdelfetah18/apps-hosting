package cmd

import (
	"go_registry/handler"
	"go_registry/storage"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Server Go Module",

	RunE: func(cmd *cobra.Command, args []string) error {
		minioStorage := storage.NewMinioStorage(false)
		requestsHandler := handler.NewHandler(minioStorage)

		router := mux.NewRouter()

		router.HandleFunc("/{base}/{module}/@v/list", requestsHandler.ListHandler)
		router.HandleFunc("/{base}/{module}/@v/{version}.info", requestsHandler.InfoHandler)
		router.HandleFunc("/{base}/{module}/@v/{version}.mod", requestsHandler.ModHandler)
		router.HandleFunc("/{base}/{module}/@v/{version}.zip", requestsHandler.ZipHandler)
		router.HandleFunc("/{base}/{module}/@latest", requestsHandler.LatestHandler)

		addr := ":8080"
		log.Printf("Server listening on %s …", addr)
		if err := http.ListenAndServe(addr, router); err != nil {
			log.Fatalf("server error: %v", err)
		}
		return nil
	},
}
