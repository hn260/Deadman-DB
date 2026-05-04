package scheduler

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
	"github.com/jmoiron/sqlx"
	"deadmandb/internal/engine"
	"deadmandb/internal/state"
	"deadmandb/internal/storage"
)

// Daemon runs the background scheduler.
type Daemon struct {
	cron    *cron.Cron
	stateDB string
	store   storage.Provider
}

func NewDaemon(baseDir string) (*Daemon, error) {
	dbPath := filepath.Join(baseDir, "state.db")
	backupsDir := filepath.Join(baseDir, "backups")
	
	store, err := storage.NewLocalProvider(backupsDir)
	if err != nil {
		return nil, err
	}

	return &Daemon{
		cron:    cron.New(cron.WithSeconds()), // Standard cron + seconds if needed, but standard is fine
		stateDB: dbPath,
		store:   store,
	}, nil
}

func (d *Daemon) Start() error {
	databases := viper.GetStringMap("databases")
	if len(databases) == 0 {
		return fmt.Errorf("no databases configured to schedule")
	}

	for dbName := range databases {
		configKey := fmt.Sprintf("databases.%s", dbName)
		schedule := viper.GetString(configKey + ".schedule")
		dbType := viper.GetString(configKey + ".type")
		dbURL := viper.GetString(configKey + ".url")

		if schedule == "" {
			log.Printf("Skipping database '%s': no schedule defined", dbName)
			continue
		}

		log.Printf("Scheduling backup for '%s' (%s) with cron: %s", dbName, dbType, schedule)
		
		// Capture variable for closure
		name, url, typ := dbName, dbURL, dbType

		_, err := d.cron.AddFunc(schedule, func() {
			log.Printf("Triggering scheduled backup for '%s'...", name)
			d.runBackup(name, typ, url)
		})

		if err != nil {
			return fmt.Errorf("failed to schedule '%s': %w", name, err)
		}
	}

	d.cron.Start()
	log.Println("Daemon started successfully. Press Ctrl+C to stop.")
	
	// Block forever (or until daemon is killed)
	select {}
}

func (d *Daemon) runBackup(dbName, dbType, dbURL string) {
	var dbEngine engine.Engine
	if dbType == "postgres" {
		dbEngine = engine.NewPostgresEngine()
	} else {
		log.Printf("Error: unsupported database type: %s", dbType)
		return
	}

	stateDB, err := state.InitDB(d.stateDB)
	if err != nil {
		log.Printf("Error initializing state DB: %v", err)
		return
	}
	defer stateDB.Close()

	snapshotID := fmt.Sprintf("snap_%d", time.Now().Unix())
	pr, pw := io.Pipe()
	ctx := context.Background()

	go func() {
		err := dbEngine.Backup(ctx, dbURL, pw)
		pw.CloseWithError(err)
	}()

	size, err := d.store.Save(ctx, snapshotID, pr)
	
	status := "success"
	if err != nil {
		status = "failed"
		log.Printf("Backup failed for '%s': %v", dbName, err)
	} else {
		log.Printf("Backup successful for '%s': saved %d bytes", dbName, size)
	}

	snapshot := state.Snapshot{
		ID:        snapshotID,
		DBName:    dbName,
		Timestamp: time.Now().Unix(),
		Size:      size,
		Status:    status,
		FilePath:  filepath.Join(filepath.Dir(d.stateDB), "backups", snapshotID+".sql.gz"),
	}

	_, err = stateDB.NamedExec(`INSERT INTO snapshots (id, db_name, timestamp, size, status, file_path) VALUES (:id, :db_name, :timestamp, :size, :status, :file_path)`, snapshot)
	if err != nil {
		log.Printf("Failed to save snapshot metadata: %v", err)
	} else if status == "success" {
		d.enforceRetentionPolicy(stateDB, dbName)
	}

	if status == "failed" {
		d.sendWebhookAlert(dbName, err)
	}
}

func (d *Daemon) sendWebhookAlert(dbName string, backupErr error) {
	webhookURL := viper.GetString("alerting.webhook_url")
	if webhookURL == "" {
		return
	}

	payload := fmt.Sprintf(`{"text": "🚨 **Deadman DB Alert** 🚨\nBackup failed for database: *%s*\nError: %v"}`, dbName, backupErr)
	
	resp, err := http.Post(webhookURL, "application/json", strings.NewReader(payload))
	if err != nil {
		log.Printf("Failed to send webhook alert: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		log.Printf("Webhook alert returned status: %d", resp.StatusCode)
	}
}

func (d *Daemon) enforceRetentionPolicy(db *sqlx.DB, dbName string) {
	configKey := fmt.Sprintf("databases.%s.keep_last", dbName)
	keepLast := viper.GetInt(configKey)
	if keepLast <= 0 {
		return // No retention policy
	}

	var snapshots []state.Snapshot
	err := db.Select(&snapshots, "SELECT * FROM snapshots WHERE db_name = ? AND status = 'success' ORDER BY timestamp DESC", dbName)
	if err != nil {
		log.Printf("Failed to fetch snapshots for retention: %v", err)
		return
	}

	if len(snapshots) <= keepLast {
		return
	}

	toDelete := snapshots[keepLast:]
	for _, snap := range toDelete {
		log.Printf("Retention policy: deleting old snapshot %s", snap.ID)
		
		// Delete from storage
		err := d.store.Delete(context.Background(), snap.ID)
		if err != nil {
			log.Printf("Failed to delete file for %s: %v", snap.ID, err)
			continue
		}

		// Delete from metadata
		_, err = db.Exec("DELETE FROM snapshots WHERE id = ?", snap.ID)
		if err != nil {
			log.Printf("Failed to delete metadata for %s: %v", snap.ID, err)
		}
	}
}
