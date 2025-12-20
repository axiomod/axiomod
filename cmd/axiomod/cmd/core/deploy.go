package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [env]",
	Short: "Deploy the application to a specified environment",
	Long: `Deploy the application to a specified environment (dev, staging, prod).

Example:
  axiomod deploy dev
  axiomod deploy prod
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := args[0]
		fmt.Printf("Deploying application to %s environment...\n", env)

		// Validate environment
		validEnvs := map[string]bool{
			"dev":        true,
			"staging":    true,
			"prod":       true,
			"production": true, // Allow full name as well
		}

		if !validEnvs[strings.ToLower(env)] {
			fmt.Printf("Invalid environment: %s. Valid options are: dev, staging, prod\n", env)
			os.Exit(1)
		}

		// In a real implementation, this would use a deployment tool like Kubernetes, Helm, etc.
		// For this example, we'll just simulate a deployment

		// First, build the Docker image
		fmt.Println("Building Docker image...")
		dockerCmd := exec.Command("docker", "build", "-t", "go-axiomod:"+env, ".")
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr

		if err := dockerCmd.Run(); err != nil {
			fmt.Printf("Docker build failed: %v\n", err)
			os.Exit(1)
		}

		// Simulate deployment
		fmt.Printf("Deploying Docker image go-axiomod:%s to %s environment...\n", env, env)

		// In a real implementation, this would push the image to a registry and deploy to the target environment
		fmt.Println("Simulating deployment steps:")
		fmt.Println("1. Pushing image to registry...")
		fmt.Println("2. Updating deployment configuration...")
		fmt.Println("3. Applying deployment to Kubernetes cluster...")
		fmt.Println("4. Waiting for deployment to complete...")

		fmt.Printf("\nApplication successfully deployed to %s environment!\n", env)
		fmt.Println("You can check the status with: axiomod status")
	},
}

// NewDeployCmd returns the deploy command.
func NewDeployCmd() *cobra.Command {
	return deployCmd
}
