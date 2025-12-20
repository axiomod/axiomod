package core

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display version information",
	Long: `Display version information for the Axiomod CLI and framework.

Example:
  axiomod version
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Axiomod Framework")
		fmt.Println("Version: 1.0.0")
		fmt.Println("Build Date: 2025-04-28")
		fmt.Println("Go Version:", runtime.Version())
		fmt.Println("OS/Arch:", runtime.GOOS+"/"+runtime.GOARCH)
	},
}

// NewVersionCmd returns the version command.
func NewVersionCmd() *cobra.Command {
	return versionCmd
}
