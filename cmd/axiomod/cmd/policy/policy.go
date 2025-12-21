package policy

import (
	"github.com/spf13/cobra"
)

// policyCmd represents the policy command
var policyCmd = &cobra.Command{
	Use:   "policy",
	Short: "Manage RBAC policies and roles",
	Long: `Manage RBAC policies and roles using Casbin.

This command has subcommands for adding, removing, and listing permissions and role assignments.

Example:
  axiomod policy add p alice data1 read
  axiomod policy add g bob admin
  axiomod policy list
`,
}

// NewPolicyCmd returns the policy command.
func NewPolicyCmd() *cobra.Command {
	return policyCmd
}
