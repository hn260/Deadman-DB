package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"deadmandb/internal/state"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Deadman DB configuration and state",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}

		baseDir := filepath.Join(homeDir, ".deadmandb")
		dbPath := filepath.Join(baseDir, "state.db")
		configPath := filepath.Join(baseDir, "config.yaml")

		fmt.Printf("Initializing Deadman DB at %s...\n", baseDir)

		// Create state DB
		_, err = state.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to initialize state DB: %w", err)
		}
		fmt.Printf("✅ Initialized state database at %s\n", dbPath)

		// Create default config file if it doesn't exist
		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			defaultConfig := `
# Deadman DB Configuration
# Add your databases and schedules here
databases:
  default:
    type: "postgres"
    url: "postgres://user:pass@localhost:5432/dbname"
    schedule: "0 2 * * *" # Every day at 2:00 AM
`
			if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
				return fmt.Errorf("failed to write config file: %w", err)
			}
			fmt.Printf("✅ Created default config at %s\n", configPath)
		} else {
			fmt.Printf("ℹ️ Config file already exists at %s\n", configPath)
		}

		fmt.Println("Initialization complete. Run `deadmandb --help` to see available commands.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
