package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"deadmandb/internal/api"
)

var port int

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the local HTTP API server for the dashboard",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		dbPath := filepath.Join(homeDir, ".deadmandb", "state.db")

		fmt.Printf("Starting API server on port %d...\n", port)
		server := api.NewServer(dbPath)
		
		return server.Start(port)
	},
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the API server on")
	rootCmd.AddCommand(serveCmd)
}
