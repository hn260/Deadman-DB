package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"deadmandb/internal/engine"
	"deadmandb/internal/state"
	"deadmandb/internal/storage"
)

var backupCmd = &cobra.Command{
	Use:   "backup [dbName]",
	Short: "Run a manual backup for a configured database",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dbName := args[0]

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

		backupsDir := filepath.Join(baseDir, "backups")
		store, err := storage.NewLocalProvider(backupsDir)
		if err != nil {
			return err
		}

		snapshotID := fmt.Sprintf("snap_%d", time.Now().Unix())
		fmt.Printf("Starting backup for '%s' (Snapshot: %s)...\n", dbName, snapshotID)

		// Create a pipe for streaming from engine to storage
		pr, pw := io.Pipe()

		ctx := context.Background()

		// Run engine backup in a goroutine
		go func() {
			err := dbEngine.Backup(ctx, dbURL, pw)
			pw.CloseWithError(err)
		}()

		// Save the snapshot in the main thread
		size, err := store.Save(ctx, snapshotID, pr)
		if err != nil {
			return fmt.Errorf("backup failed: %w", err)
		}

		// Save metadata to state DB
		snapshot := state.Snapshot{
			ID:        snapshotID,
			DBName:    dbName,
			Timestamp: time.Now().Unix(),
			Size:      size,
			Status:    "success",
			FilePath:  filepath.Join(backupsDir, snapshotID+".sql.gz"),
		}

		_, err = stateDB.NamedExec(`INSERT INTO snapshots (id, db_name, timestamp, size, status, file_path) VALUES (:id, :db_name, :timestamp, :size, :status, :file_path)`, snapshot)
		if err != nil {
			return fmt.Errorf("failed to save snapshot metadata: %w", err)
		}

		fmt.Printf("✅ Backup complete! Saved %d bytes.\n", size)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)
}
