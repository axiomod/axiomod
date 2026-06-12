package core

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var (
	deployDryRun    bool
	deployRegistry  string
	deployManifests string
)

// deployCmd represents the deploy command
var deployCmd = &cobra.Command{
	Use:   "deploy [env]",
	Short: "Build and (optionally) push/deploy the application",
	Long: `Build the application's Docker image for an environment.

By default this is a DRY RUN: the image is built locally and the remaining
steps are printed as [SIMULATED]. Pass --dry-run=false together with
--registry (for the push) and --manifests (for kubectl apply) to execute
them for real.

Example:
  axiomod deploy dev
  axiomod deploy prod --dry-run=false --registry registry.example.com/team --manifests deploy/kubernetes
`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		env := strings.ToLower(args[0])

		validEnvs := map[string]bool{"dev": true, "staging": true, "prod": true, "production": true}
		if !validEnvs[env] {
			fmt.Printf("Invalid environment: %s. Valid options are: dev, staging, prod\n", env)
			os.Exit(1)
		}

		module, err := moduleName()
		if err != nil {
			fmt.Printf("Error reading go.mod: %v\n", err)
			os.Exit(1)
		}
		image := fmt.Sprintf("%s:%s", module, env)

		fmt.Printf("Building Docker image %s ...\n", image)
		dockerCmd := exec.Command("docker", "build", "-t", image, ".")
		dockerCmd.Stdout = os.Stdout
		dockerCmd.Stderr = os.Stderr
		if err := dockerCmd.Run(); err != nil {
			fmt.Printf("Docker build failed: %v\n", err)
			os.Exit(1)
		}

		if deployDryRun {
			fmt.Println("\nDry run (default) — the following steps were NOT executed:")
			fmt.Printf("  [SIMULATED] docker tag %s <registry>/%s && docker push\n", image, image)
			fmt.Println("  [SIMULATED] kubectl apply -f <manifests>")
			fmt.Println("\nPass --dry-run=false with --registry and --manifests to execute them.")
			return
		}

		if deployRegistry == "" {
			fmt.Println("Error: --dry-run=false requires --registry for the image push.")
			os.Exit(1)
		}

		remote := strings.TrimSuffix(deployRegistry, "/") + "/" + image
		fmt.Printf("\nPushing %s ...\n", remote)
		for _, c := range [][]string{
			{"docker", "tag", image, remote},
			{"docker", "push", remote},
		} {
			run := exec.Command(c[0], c[1:]...)
			run.Stdout = os.Stdout
			run.Stderr = os.Stderr
			if err := run.Run(); err != nil {
				fmt.Printf("%s failed: %v\n", strings.Join(c, " "), err)
				os.Exit(1)
			}
		}

		if deployManifests == "" {
			fmt.Println("\nImage pushed. No --manifests given, skipping kubectl apply.")
			return
		}

		fmt.Printf("\nApplying manifests from %s ...\n", deployManifests)
		kubectl := exec.Command("kubectl", "apply", "-f", deployManifests)
		kubectl.Stdout = os.Stdout
		kubectl.Stderr = os.Stderr
		if err := kubectl.Run(); err != nil {
			fmt.Printf("kubectl apply failed: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("\nApplication deployed to %s.\n", env)
		fmt.Println("You can check the status with: axiomod status")
	},
}

// NewDeployCmd returns the deploy command.
func NewDeployCmd() *cobra.Command {
	deployCmd.Flags().BoolVar(&deployDryRun, "dry-run", true,
		"build only; print push/apply steps as [SIMULATED]")
	deployCmd.Flags().StringVar(&deployRegistry, "registry", "",
		"container registry prefix for the push (required with --dry-run=false)")
	deployCmd.Flags().StringVar(&deployManifests, "manifests", "",
		"path to Kubernetes manifests for kubectl apply")
	return deployCmd
}
