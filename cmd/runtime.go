package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/nuonco/nuon/bins/cli/extensions/nuon-ext-terraform/internal/config"
	"github.com/nuonco/nuon/sdks/nuon-go"
)

type runtimeState struct {
	cfg *config.Runtime
	api nuon.Client
}

func newRuntime() (*runtimeState, error) {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		return nil, err
	}

	apiClient, err := nuon.New(
		nuon.WithURL(cfg.APIURL),
		nuon.WithAuthToken(cfg.APIToken),
		nuon.WithOrgID(cfg.OrgID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to initialize api client: %w", err)
	}

	return &runtimeState{
		cfg: cfg,
		api: apiClient,
	}, nil
}

func (r *runtimeState) resolveInstallID(override string) (string, error) {
	if override != "" {
		return override, nil
	}
	if r.cfg.InstallID != "" {
		return r.cfg.InstallID, nil
	}
	return "", fmt.Errorf("install id is required: set NUON_INSTALL_ID or pass --install-id")
}

func withRuntime(run func(*runtimeState, *cobra.Command, []string) error) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		rt, err := newRuntime()
		if err != nil {
			return err
		}
		return run(rt, cmd, args)
	}
}
