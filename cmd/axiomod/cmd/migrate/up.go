package migrate

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"

	// Import necessary database drivers
	_ "github.com/lib/pq" // PostgreSQL driver
	// _ "github.com/go-sql-driver/mysql" // MySQL driver (if needed)
)

// upCmd represents the migrate up command
var upCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending database migrations",
	Long: `Apply all pending database migrations found in the migrations directory.

Example:
  axiomod migrate up
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Applying pending migrations...")

		// Load configuration to get DB DSN
		// TODO: Integrate with centralized config loading
		dbDSN := os.Getenv("DATABASE_DSN") // Example: Get DSN from env var
		if dbDSN == "" {
			// Fallback or load from config file
			// cfg, err := config.Load("") // Assuming config loading logic
			// if err != nil {
			// 	 fmt.Printf("Error loading config: %v\n", err)
			// 	 os.Exit(1)
			// }
			// dbDSN = cfg.Database.DSN
			dbDSN = "postgres://postgres:postgres@localhost:5432/axiomod?sslmode=disable" // Default for example
			fmt.Println("Warning: DATABASE_DSN not set, using default DSN.")
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

		// Apply migrations
		err = m.Up()
		if err != nil {
			if err == migrate.ErrNoChange {
				fmt.Println("No new migrations to apply.")
			} else {
				fmt.Printf("Error applying migrations: %v\n", err)
				os.Exit(1)
			}
		} else {
			fmt.Println("Migrations applied successfully.")
		}
	},
}

// Removed ensureDatabaseExists function from here, moved to utils.go

// NewUpCmd returns the migrate up command.
func NewUpCmd() *cobra.Command {
	return upCmd
}

func init() {
	// Add subcommands to the parent migrateCmd
	migrateCmd.AddCommand(upCmd)
}
