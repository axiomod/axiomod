package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var statusBaseURL string

// healthComponent mirrors framework/health.Component for decoding.
type healthComponent struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// healthResponse mirrors framework/health.Response for decoding.
type healthResponse struct {
	Status     string                     `json:"status"`
	Components map[string]healthComponent `json:"components"`
	Timestamp  string                     `json:"timestamp"`
}

// statusCmd represents the status command
var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the status of the running application",
	Long: `Check the status of the running Axiomod application by querying its
readiness endpoint and reporting per-component health.

Example:
  axiomod status
  axiomod status --url http://localhost:8080
`,
	Run: func(cmd *cobra.Command, args []string) {
		base := strings.TrimSuffix(statusBaseURL, "/")
		fmt.Printf("Checking application status at %s ...\n", base)

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(base + "/ready")
		if err != nil {
			fmt.Printf("Application is not reachable: %v\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		var health healthResponse
		if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
			fmt.Printf("Could not decode health response (HTTP %d): %v\n", resp.StatusCode, err)
			os.Exit(1)
		}

		fmt.Printf("\nStatus: %s (HTTP %d)\n", health.Status, resp.StatusCode)
		if len(health.Components) > 0 {
			fmt.Println("\nComponents:")
			for name, c := range health.Components {
				line := fmt.Sprintf("  - %-20s %s", name, c.Status)
				if c.Error != "" {
					line += "  (" + c.Error + ")"
				}
				fmt.Println(line)
			}
		}

		if resp.StatusCode != http.StatusOK || !strings.EqualFold(health.Status, "UP") {
			fmt.Println("\nApplication is not healthy.")
			os.Exit(1)
		}
		fmt.Println("\nApplication is running and healthy.")
	},
}

// NewStatusCmd returns the status command.
func NewStatusCmd() *cobra.Command {
	statusCmd.Flags().StringVar(&statusBaseURL, "url", "http://localhost:8080",
		"base URL of the running application")
	return statusCmd
}
