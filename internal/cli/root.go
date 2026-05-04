package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "deadmandb",
	Short: "Deadman DB is an automated database backup and recovery system",
	Long:  `Deadman DB continuously snapshots databases, stores versioned backups, and enables fast, reliable restoration.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		return err
	}
	return nil
}

func init() {
	// Add global flags here if needed
}
