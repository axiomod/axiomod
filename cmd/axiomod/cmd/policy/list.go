package policy

import (
	"fmt"
	"os"

	"github.com/axiomod/axiomod/framework/auth"
	"github.com/axiomod/axiomod/framework/config"
	"github.com/spf13/cobra"
)

// listCmd represents the policy list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all RBAC policies",
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

		enforcer := rbac.GetEnforcer()

		fmt.Println("Grouping Policies (Roles Assignment):")
		groupingPolicies, _ := enforcer.GetGroupingPolicy()
		for _, p := range groupingPolicies {
			fmt.Printf("  g, %v\n", p)
		}

		fmt.Println("\nPolicies (Permissions):")
		policies, _ := enforcer.GetPolicy()
		for _, p := range policies {
			fmt.Printf("  p, %v\n", p)
		}
	},
}

func init() {
	policyCmd.AddCommand(listCmd)
}
