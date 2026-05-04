package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"deadmandb/internal/scheduler"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Start the background scheduler to run automated backups",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		baseDir := filepath.Join(homeDir, ".deadmandb")
		configPath := filepath.Join(baseDir, "config.yaml")

		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config (did you run init?): %w", err)
		}

		d, err := scheduler.NewDaemon(baseDir)
		if err != nil {
			return fmt.Errorf("failed to initialize daemon: %w", err)
		}

		return d.Start()
	},
}

func init() {
	rootCmd.AddCommand(daemonCmd)
}
