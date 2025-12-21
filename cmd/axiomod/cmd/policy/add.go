package policy

import (
	"fmt"
	"os"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/spf13/cobra"
)

// addCmd represents the policy add command
var addCmd = &cobra.Command{
	Use:   "add [p|g] [args...]",
	Short: "Add a policy or grouping policy",
	Long: `Add a policy (permission) or a grouping policy (role assignment).

Usage:
  axiomod policy add p <sub> <obj> <act>
  axiomod policy add g <user> <role>
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
			ok, err = rbac.AddPolicy(args[1], args[2], args[3])
		case "g":
			if len(args) != 3 {
				fmt.Println("Error: 'g' policy requires 2 arguments: user, role")
				os.Exit(1)
			}
			ok, err = rbac.AddRoleForUser(args[1], args[2])
		default:
			fmt.Printf("Error: unknown policy type '%s'. Use 'p' or 'g'.\n", args[0])
			os.Exit(1)
		}

		if err != nil {
			fmt.Printf("Error adding policy: %v\n", err)
			os.Exit(1)
		}

		if ok {
			fmt.Println("Policy added successfully.")
			// For CSV file persistence, Casbin often needs SavePolicy if not auto-save.
			// The default file adapter in casbin/v2 doesn't always auto-save on mutation
			// unless we explicitly call SavePolicy.
			if err := rbac.GetEnforcer().SavePolicy(); err != nil {
				fmt.Printf("Warning: failed to save policy to storage: %v\n", err)
			}
		} else {
			fmt.Println("Policy already exists.")
		}
	},
}

func init() {
	policyCmd.AddCommand(addCmd)
}
