package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon-ext-terraform/internal/docker"
	"github.com/nuonco/nuon-ext-terraform/internal/resolve"
)

func newComponentCmd() *cobra.Command {
	var installID string
	var componentName string
	var sourceDir string

	run := func(rt *runtimeState, cmd *cobra.Command, args []string) error {
		resolvedInstallID, err := rt.resolveInstallID(installID)
		if err != nil {
			return err
		}

		if componentName == "" {
			names, listErr := resolve.ListTerraformComponents(cmd.Context(), rt.api, resolvedInstallID)
			if listErr == nil && len(names) > 0 {
				return fmt.Errorf("--name is required; available terraform components: %s", strings.Join(names, ", "))
			}
			return fmt.Errorf("--name is required")
		}

		if len(args) == 0 {
			return fmt.Errorf("specify a terraform command, e.g.: init, plan, apply, state list, shell")
		}

		ctx := cmd.Context()

		fmt.Fprintf(os.Stderr, "Resolving component %q workspace for install %s...\n", componentName, resolvedInstallID)
		info, err := resolve.ComponentWorkspace(ctx, rt.api, resolvedInstallID, componentName)
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
	}

	cmd := &cobra.Command{
		Use:   "component [terraform args...]",
		Short: "Run Terraform commands against a component workspace",
		Long: `Resolve a Terraform component's workspace and run a Terraform command inside
a Docker container. The container uses the exact Terraform version configured
for the component.

For public repos, the source is cloned automatically. For connected repos,
provide the local source directory with --source-dir.

Use the "shell" subcommand to open an interactive shell in the container with
the Nuon backend pre-configured.`,
		Example: strings.Join([]string{
			"  nuon terraform component init --name certificate_wildcard_public",
			"  nuon terraform component plan --name certificate_wildcard_public",
			"  nuon terraform component apply --name my-database",
			"  nuon terraform component state list --name my-database",
			"  nuon terraform component shell --name my-database",
		}, "\n"),
		Args:               cobra.ArbitraryArgs,
		DisableFlagParsing: false,
		RunE:               withRuntime(run),
	}

	cmd.PersistentFlags().StringVar(&installID, "install-id", "", "Install ID (defaults to NUON_INSTALL_ID)")
	cmd.PersistentFlags().StringVar(&componentName, "name", "", "Component name (required)")
	cmd.PersistentFlags().StringVar(&sourceDir, "source-dir", "", "Local directory containing Terraform source (required for connected repos)")

	shellCmd := &cobra.Command{
		Use:   "shell",
		Short: "Open an interactive shell in the component's Terraform container",
		Long: `Open an interactive /bin/sh shell inside the Terraform container for the
component, with the Nuon HTTP backend pre-configured at /workspace/nuon_backend.tf.

Useful for ad-hoc terraform commands, debugging state, or running providers
against the same backend the platform uses.`,
		Example: "  nuon terraform component shell --name my-database",
		Args:    cobra.NoArgs,
		RunE: withRuntime(func(rt *runtimeState, cmd *cobra.Command, args []string) error {
			return run(rt, cmd, []string{"shell"})
		}),
	}
	cmd.AddCommand(shellCmd)

	return cmd
}
