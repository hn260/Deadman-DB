package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"deadmandb/internal/engine"
	"deadmandb/internal/state"
	"deadmandb/internal/storage"
)

var snapshotID string

var restoreCmd = &cobra.Command{
	Use:   "restore [dbName]",
	Short: "Restore a database from a snapshot",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbName := args[0]

		if snapshotID == "" {
			return fmt.Errorf("must provide --snapshot-id")
		}

		// Prompt for confirmation
		fmt.Printf("⚠️ WARNING: Restoring will overwrite existing data in '%s'.\n", dbName)
		fmt.Printf("Are you sure you want to restore snapshot %s? (y/N): ", snapshotID)
		
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return err
		}

		response = strings.ToLower(strings.TrimSpace(response))
		if response != "y" && response != "yes" {
			fmt.Println("Restore cancelled.")
			return nil
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		baseDir := filepath.Join(homeDir, ".deadmandb")
		dbPath := filepath.Join(baseDir, "state.db")
		configPath := filepath.Join(baseDir, "config.yaml")

		viper.SetConfigFile(configPath)
		if err := viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config: %w", err)
		}

		// Get db config
		configKey := fmt.Sprintf("databases.%s", dbName)
		if !viper.IsSet(configKey) {
			return fmt.Errorf("database '%s' not found in config", dbName)
		}

		dbType := viper.GetString(configKey + ".type")
		dbURL := viper.GetString(configKey + ".url")

		var dbEngine engine.Engine
		if dbType == "postgres" {
			dbEngine = engine.NewPostgresEngine()
		} else {
			return fmt.Errorf("unsupported database type: %s", dbType)
		}

		stateDB, err := state.InitDB(dbPath)
		if err != nil {
			return err
		}
		defer stateDB.Close()

		// Verify snapshot exists in DB
		var snap state.Snapshot
		err = stateDB.Get(&snap, "SELECT * FROM snapshots WHERE id = ?", snapshotID)
		if err != nil {
			return fmt.Errorf("snapshot %s not found in metadata: %w", snapshotID, err)
		}

		backupsDir := filepath.Join(baseDir, "backups")
		store, err := storage.NewLocalProvider(backupsDir)
		if err != nil {
			return err
		}

		fmt.Printf("Restoring snapshot %s to '%s'...\n", snapshotID, dbName)

		ctx := context.Background()

		// Retrieve uncompressed stream from storage
		rc, err := store.Retrieve(ctx, snapshotID)
		if err != nil {
			return fmt.Errorf("failed to read snapshot data: %w", err)
		}
		defer rc.Close()

		// Pipe the uncompressed data into the restore engine
		err = dbEngine.Restore(ctx, dbURL, rc)
		if err != nil {
			return fmt.Errorf("restore failed: %w", err)
		}

		fmt.Println("✅ Restore completed successfully!")
		return nil
	},
}

func init() {
	restoreCmd.Flags().StringVarP(&snapshotID, "snapshot-id", "s", "", "ID of the snapshot to restore")
	restoreCmd.MarkFlagRequired("snapshot-id")
	rootCmd.AddCommand(restoreCmd)
}
