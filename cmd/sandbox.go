package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/extensions/nuon-ext-terraform/internal/docker"
	"github.com/nuonco/nuon/bins/cli/extensions/nuon-ext-terraform/internal/resolve"
)

func newSandboxCmd() *cobra.Command {
	var installID string
	var sourceDir string

	cmd := &cobra.Command{
		Use:   "sandbox [terraform args...]",
		Short: "Run Terraform commands against a sandbox workspace",
		Long: `Resolve the sandbox's Terraform workspace and run a Terraform command inside
a Docker container. The container uses the exact Terraform version configured
for the sandbox.

For public repos, the source is cloned automatically. For connected repos,
provide the local source directory with --source-dir.`,
		Example: strings.Join([]string{
			"  nuon terraform sandbox init",
			"  nuon terraform sandbox plan",
			"  nuon terraform sandbox apply",
			"  nuon terraform sandbox state list",
			"  nuon terraform sandbox state show aws_eks_cluster.main",
			"  nuon terraform sandbox shell",
			"  nuon terraform sandbox init --source-dir ~/code/my-sandbox",
		}, "\n"),
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: false,
		RunE: withRuntime(func(rt *runtimeState, cmd *cobra.Command, args []string) error {
			resolvedInstallID, err := rt.resolveInstallID(installID)
			if err != nil {
				return err
			}

			if len(args) == 0 {
				return fmt.Errorf("specify a terraform command, e.g.: init, plan, apply, state list")
			}

			ctx := cmd.Context()

			fmt.Fprintf(os.Stderr, "Resolving sandbox workspace for install %s...\n", resolvedInstallID)
			info, err := resolve.SandboxWorkspace(ctx, rt.api, resolvedInstallID)
			if err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Workspace:  %s\n", info.WorkspaceID)
			fmt.Fprintf(os.Stderr, "Terraform:  v%s\n", info.TerraformVersion)

			src, cleanupDir, err := resolveSource(info, sourceDir)
			if err != nil {
				return err
			}
			if cleanupDir != "" {
				defer os.RemoveAll(cleanupDir)
			}

			if err := docker.EnsureImage(info.TerraformVersion); err != nil {
				return fmt.Errorf("unable to pull terraform image: %w", err)
			}

			cfg := docker.RunConfig{
				TerraformVersion: info.TerraformVersion,
				WorkspaceID:      info.WorkspaceID,
				OrgID:            rt.cfg.OrgID,
				APIURL:           rt.cfg.APIURL,
				APIToken:         rt.cfg.APIToken,
				SourceDir:        src,
				Command:          args,
			}

			if args[0] == "init" {
				return docker.RunInit(cfg)
			}

			return docker.Run(cfg)
		}),
	}

	cmd.Flags().StringVar(&installID, "install-id", "", "Install ID (defaults to NUON_INSTALL_ID)")
	cmd.Flags().StringVar(&sourceDir, "source-dir", "", "Local directory containing Terraform source (required for connected repos)")

	return cmd
}
