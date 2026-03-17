package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "nuon-ext-terraform",
		Short: "Run Terraform CLI commands against Nuon-managed workspaces",
		Long: `Run Terraform commands inside a Docker container connected to a Nuon-managed
sandbox or component workspace. The extension automatically resolves the
workspace, Terraform version, and source code from the API.

Use the "sandbox" or "component" subcommands to target a specific workspace.`,
		Example: strings.Join([]string{
			"  nuon terraform sandbox init",
			"  nuon terraform sandbox plan",
			"  nuon terraform sandbox state list",
			"  nuon terraform component init --name certificate_wildcard_public",
			"  nuon terraform component plan --name certificate_wildcard_public",
			"  nuon terraform component state show aws_acm_certificate.cert --name my-component",
		}, "\n"),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.AddCommand(
		newSandboxCmd(),
		newComponentCmd(),
	)

	return cmd
}
