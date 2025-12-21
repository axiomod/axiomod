package migrate

import (
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/spf13/cobra"
)

// versionCmd represents the migrate version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the current migration version",
	Long: `Print the current migration version of the database.

Example:
  axiomod migrate version
`,
	Run: func(cmd *cobra.Command, args []string) {
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

		version, dirty, err := m.Version()
		if err != nil {
			if err == migrate.ErrNilVersion {
				fmt.Println("No migrations have been applied yet.")
				return
			}
			fmt.Printf("Error getting version: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Current version: %d (dirty: %v)\n", version, dirty)
	},
}

func init() {
	migrateCmd.AddCommand(versionCmd)
}
