package core

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
)

// healthcheckCmd represents the healthcheck command
var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Check the health of the running application",
	Long: `Perform a health check on the running Go Macroservice application.

This command sends a request to the /health endpoint.

Example:
  axiomod healthcheck
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Performing health check...")

		// TODO: Make the health check URL configurable
		healthURL := "http://localhost:8080/health"

		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		resp, err := client.Get(healthURL)
		if err != nil {
			fmt.Printf("Health check failed: %v\n", err)
			// Consider exiting with a non-zero code if needed
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			fmt.Printf("Application is healthy (Status: %s)\n", resp.Status)
		} else {
			fmt.Printf("Application is unhealthy (Status: %s)\n", resp.Status)
			// Consider exiting with a non-zero code if needed
		}
	},
}

// NewHealthcheckCmd returns the healthcheck command.
func NewHealthcheckCmd() *cobra.Command {
	return healthcheckCmd
}
