package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"deadmandb/internal/state"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available snapshots",
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}

		dbPath := filepath.Join(homeDir, ".deadmandb", "state.db")

		stateDB, err := state.InitDB(dbPath)
		if err != nil {
			return fmt.Errorf("failed to connect to state database: %w", err)
		}
		defer stateDB.Close()

		var snapshots []state.Snapshot
		err = stateDB.Select(&snapshots, "SELECT * FROM snapshots ORDER BY timestamp DESC")
		if err != nil {
			return fmt.Errorf("failed to fetch snapshots: %w", err)
		}

		if len(snapshots) == 0 {
			fmt.Println("No snapshots found.")
			return nil
		}

		// Use tabwriter for nice columnar output
		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "SNAPSHOT ID\tDATABASE\tDATE\tSIZE (BYTES)\tSTATUS")

		for _, s := range snapshots {
			dateStr := time.Unix(s.Timestamp, 0).Format(time.RFC33Context)
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%s\n", s.ID, s.DBName, dateStr, s.Size, s.Status)
		}

		w.Flush()
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
