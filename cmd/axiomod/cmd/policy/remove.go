package policy

import (
	"fmt"
	"os"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/spf13/cobra"
)

// removeCmd represents the policy remove command
var removeCmd = &cobra.Command{
	Use:   "remove [p|g] [args...]",
	Short: "Remove a policy or grouping policy",
	Long: `Remove a policy (permission) or a grouping policy (role assignment).

Usage:
  axiomod policy remove p <sub> <obj> <act>
  axiomod policy remove g <user> <role>
`,
	Args: cobra.MinimumNArgs(3),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load("")
		if err != nil {
			fmt.Printf("Error loading config: %v\n", err)
			os.Exit(1)
		}

		rbac, err := auth.NewRBACService(cfg.Casbin)
		if err != nil {
			fmt.Printf("Error initializing RBAC service: %v\n", err)
			os.Exit(1)
		}

		var ok bool
		switch args[0] {
		case "p":
			if len(args) != 4 {
				fmt.Println("Error: 'p' policy requires 3 arguments: sub, obj, act")
				os.Exit(1)
			}
			ok, err = rbac.RemovePolicy(args[1], args[2], args[3])
		case "g":
			if len(args) != 3 {
				fmt.Println("Error: 'g' policy requires 2 arguments: user, role")
				os.Exit(1)
			}
			ok, err = rbac.RemoveRoleForUser(args[1], args[2])
		default:
			fmt.Printf("Error: unknown policy type '%s'. Use 'p' or 'g'.\n", args[0])
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("Error removing policy: %v\n", err)
			os.Exit(1)
		}

		if ok {
			fmt.Println("Policy removed successfully.")
			if err := rbac.GetEnforcer().SavePolicy(); err != nil {
				fmt.Printf("Warning: failed to save policy to storage: %v\n", err)
			}
		} else {
			fmt.Println("Policy not found.")
		}
	},
}

func init() {
	policyCmd.AddCommand(removeCmd)
}
