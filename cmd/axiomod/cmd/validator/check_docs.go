package validator

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// checkDocsCmd represents the validator check-docs command
var checkDocsCmd = &cobra.Command{
	Use:   "check-docs",
	Short: "Check if code changes have documentation updates",
	Long: `Check if recent code changes have corresponding documentation updates.

This validator helps ensure that documentation stays in sync with code changes.

Example:
  axiomod validator check-docs
  axiomod validator check-docs --since=HEAD~5
`,
	Run: func(cmd *cobra.Command, args []string) {
		since, _ := cmd.Flags().GetString("since")
		fmt.Println("Checking documentation updates...")

		if since == "" {
			since = "HEAD~1" // Default to check against previous commit
		}

		fmt.Printf("Checking changes since: %s\n", since)

		// Check if git is available
		_, err := exec.LookPath("git")
		if err != nil {
			fmt.Println("Error: git command not found. This validator requires git.")
			os.Exit(1)
		}

		// Get changed files
		changedFilesCmd := exec.Command("git", "diff", "--name-only", since)
		changedFilesOutput, err := changedFilesCmd.Output()
		if err != nil {
			fmt.Printf("Error getting changed files: %v\n", err)
			os.Exit(1)
		}

		changedFiles := strings.Split(string(changedFilesOutput), "\n")

		// Filter for code files that might need documentation
		codeFiles := []string{}
		docFiles := []string{}

		for _, file := range changedFiles {
			if file == "" {
				continue
			}

			ext := filepath.Ext(file)
			if ext == ".go" {
				codeFiles = append(codeFiles, file)
			} else if ext == ".md" || strings.Contains(file, "docs/") {
				docFiles = append(docFiles, file)
			}
		}

		if len(codeFiles) == 0 {
			fmt.Println("No code changes detected.")
			return
		}

		fmt.Printf("Found %d changed code files and %d changed documentation files.\n",
			len(codeFiles), len(docFiles))

		// Simple heuristic: if code files changed but no doc files changed, warn
		if len(docFiles) == 0 {
			fmt.Println("\nWARNING: Code changes detected but no documentation updates found.")
			fmt.Println("Consider updating relevant documentation for:")
			for _, file := range codeFiles {
				fmt.Printf("- %s\n", file)
			}
			os.Exit(1)
		}

		fmt.Println("Documentation check passed.")
	},
}

// NewCheckDocsCmd returns the validator check-docs command.
func NewCheckDocsCmd() *cobra.Command {
	checkDocsCmd.Flags().StringP("since", "s", "", "Git reference to check changes since (default: HEAD~1)")
	return checkDocsCmd
}

func init() {
	// Add subcommands to the parent validatorCmd
	validatorCmd.AddCommand(checkDocsCmd)
}
