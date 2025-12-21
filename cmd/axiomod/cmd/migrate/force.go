package migrate

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

// forceCmd represents the migrate force command
var forceCmd = &cobra.Command{
	Use:   "force [version]",
	Short: "Force the migration version",
	Long: `Force the migration version in the schema_migrations table.
This is useful for fixing a dirty migration state.

Example:
  axiomod migrate force 20230101000000
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		version, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Println("Invalid version. Please provide an integer version.")
			os.Exit(1)
		}

		fmt.Printf("Forcing migration version to: %d\n", version)

		dbDSN, err := getDSN()
		if err != nil {
			fmt.Printf("Error getting database connection string: %v\n", err)
			os.Exit(1)
		}

		m, err := migrate.New("file://migrations", dbDSN)
		if err != nil {
			fmt.Printf("Error creating migration instance: %v\n", err)
			os.Exit(1)
		}
		defer m.Close()

		if err := m.Force(version); err != nil {
			fmt.Printf("Error forcing version: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Successfully forced migration version.")
	},
}

func init() {
	migrateCmd.AddCommand(forceCmd)
}
