package migrate

import (
	"github.com/spf13/cobra"
)

// migrateCmd represents the migrate command
var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage database migrations",
	Long: `Manage database migrations for the Go Macroservice project.

This command has subcommands for creating, running, and rolling back migrations.

Example:
  axiomod migrate create add_users_table
  axiomod migrate up
  axiomod migrate down
`,
}

// NewMigrateCmd returns the migrate command.
func NewMigrateCmd() *cobra.Command {
	return migrateCmd
}
