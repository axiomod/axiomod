package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
)

// createCmd represents the migrate create command
var createCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new migration file",
	Long: `Create a new migration file with a timestamp and the provided name.

Example:
  axiomod migrate create add_users_table
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		migrationName := args[0]
		fmt.Printf("Creating migration: %s\n", migrationName)

		// Create migrations directory if it doesn't exist
		migrationsDir := "migrations"
		if err := os.MkdirAll(migrationsDir, 0755); err != nil {
			fmt.Printf("Error creating migrations directory: %v\n", err)
			os.Exit(1)
		}

		// Generate timestamp
		timestamp := time.Now().Format("20060102150405")

		// Create migration file names
		upFileName := fmt.Sprintf("%s_%s.up.sql", timestamp, migrationName)
		downFileName := fmt.Sprintf("%s_%s.down.sql", timestamp, migrationName)

		// Create up migration file
		upFilePath := filepath.Join(migrationsDir, upFileName)
		upContent := fmt.Sprintf("-- Migration: %s (up)\n\n-- Write your UP migration SQL here\n\n", migrationName)
		if err := os.WriteFile(upFilePath, []byte(upContent), 0644); err != nil {
			fmt.Printf("Error creating up migration file: %v\n", err)
			os.Exit(1)
		}

		// Create down migration file
		downFilePath := filepath.Join(migrationsDir, downFileName)
		downContent := fmt.Sprintf("-- Migration: %s (down)\n\n-- Write your DOWN migration SQL here\n\n", migrationName)
		if err := os.WriteFile(downFilePath, []byte(downContent), 0644); err != nil {
			fmt.Printf("Error creating down migration file: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Created migration files:\n- %s\n- %s\n", upFilePath, downFilePath)
	},
}

// NewCreateCmd returns the migrate create command.
func NewCreateCmd() *cobra.Command {
	return createCmd
}

func init() {
	// Add subcommands to the parent migrateCmd
	migrateCmd.AddCommand(createCmd)
}
