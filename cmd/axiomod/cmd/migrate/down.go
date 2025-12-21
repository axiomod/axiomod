package migrate

import (
	"fmt"
	"os"
	"strconv" // Added missing import

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	// Import necessary database drivers
	_ "github.com/lib/pq" // PostgreSQL driver
	// _ "github.com/go-sql-driver/mysql" // MySQL driver (if needed)
)

// downCmd represents the migrate down command
var downCmd = &cobra.Command{
	Use:   "down [N]",
	Short: "Roll back the last N migrations",
	Long: `Roll back the last N applied database migrations.
If N is not specified, it defaults to rolling back the last 1 migration.

Example:
  axiomod migrate down    # Roll back 1 migration
  axiomod migrate down 2  # Roll back 2 migrations
`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		steps := 1 // Default steps to rollback
		var err error
		if len(args) == 1 {
			steps, err = strconv.Atoi(args[0])
			if err != nil || steps <= 0 {
				fmt.Println("Invalid number of steps. Please provide a positive integer.")
				os.Exit(1)
			}
		}

		fmt.Printf("Rolling back last %d migration(s)...\n", steps)

		// Load configuration to get DB DSN
		// Load configuration to get DB DSN
		dbDSN, err := getDSN()
		if err != nil {
			fmt.Printf("Error getting database connection string: %v\n", err)
			os.Exit(1)
		}

		// Ensure the database exists (optional, depends on workflow)
		if err := ensureDatabaseExists(dbDSN); err != nil {
			fmt.Printf("Error ensuring database exists: %v\n", err)
			// Decide if this is a fatal error
		}

		// Create migrate instance
		m, err := migrate.New(
			"file://migrations", // Source URL for migration files
			dbDSN,               // Database URL
		)
		if err != nil {
			fmt.Printf("Error creating migration instance: %v\n", err)
			os.Exit(1)
		}
		defer m.Close()

		// Roll back migrations
		err = m.Steps(-steps) // Negative steps for rollback
		if err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("No migrations to roll back.")
			} else if err == migrate.ErrNilVersion {
				fmt.Println("No migrations have been applied yet.")
			} else {
				fmt.Printf("Error rolling back migrations: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Printf("Successfully rolled back %d migration(s).\n", steps)
		}
	},
}

// NewDownCmd returns the migrate down command.
func NewDownCmd() *cobra.Command {
	return downCmd
}

// Removed ensureDatabaseExists function from here, moved to utils.go

func init() {
	// Add subcommands to the parent migrateCmd
	migrateCmd.AddCommand(downCmd)
}
