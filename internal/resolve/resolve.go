package resolve

import (
	"context"
	"fmt"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type VCSSource struct {
	Repo      string
	Directory string
	Branch    string
	IsPublic  bool
}

type WorkspaceInfo struct {
	WorkspaceID      string
	TerraformVersion string
	VCS              *VCSSource
	Name             string
}

func SandboxWorkspace(ctx context.Context, api nuon.Client, installID string) (*WorkspaceInfo, error) {
	install, err := api.GetInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	if install.Sandbox == nil {
		return nil, fmt.Errorf("install %s has no sandbox", installID)
	}
	if install.Sandbox.TerraformWorkspace == nil {
		return nil, fmt.Errorf("sandbox has no terraform workspace")
	}

	appID := install.AppID
	appConfigID := install.AppConfigID
	if appID == "" || appConfigID == "" {
		return nil, fmt.Errorf("install is missing app_id or app_config_id")
	}

	recurse := true
	appConfig, err := api.GetAppConfig(ctx, appID, appConfigID, &recurse)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}

	if appConfig.Sandbox == nil {
		return nil, fmt.Errorf("app config has no sandbox configuration")
	}

	info := &WorkspaceInfo{
		WorkspaceID:      install.Sandbox.TerraformWorkspace.ID,
		TerraformVersion: appConfig.Sandbox.TerraformVersion,
		Name:             "sandbox",
	}

	info.VCS = extractSandboxVCS(appConfig.Sandbox)
	return info, nil
}

func ComponentWorkspace(ctx context.Context, api nuon.Client, installID, componentName string) (*WorkspaceInfo, error) {
	install, err := api.GetInstall(ctx, installID)
	if err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	components, _, err := api.GetInstallComponents(ctx, installID, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}

	var target *models.AppInstallComponent
	for _, c := range components {
		if c.Component != nil && strings.EqualFold(c.Component.Name, componentName) {
			target = c
			break
		}
	}
	if target == nil {
		available := make([]string, 0)
		for _, c := range components {
			if c.Component != nil && c.Component.Type == "terraform_module" {
				available = append(available, c.Component.Name)
			}
		}
		return nil, fmt.Errorf("component %q not found; available terraform components: %s", componentName, strings.Join(available, ", "))
	}

	if target.Component.Type != "terraform_module" {
		return nil, fmt.Errorf("component %q is type %q, not terraform_module", componentName, target.Component.Type)
	}

	if target.TerraformWorkspace == nil {
		return nil, fmt.Errorf("component %q has no terraform workspace", componentName)
	}

	appID := install.AppID
	appConfigID := install.AppConfigID
	if appID == "" || appConfigID == "" {
		return nil, fmt.Errorf("install is missing app_id or app_config_id")
	}

	recurse := true
	appConfig, err := api.GetAppConfig(ctx, appID, appConfigID, &recurse)
	if err != nil {
		return nil, fmt.Errorf("unable to get app config: %w", err)
	}

	var tfConfig *models.AppComponentConfigConnection
	for _, cc := range appConfig.ComponentConfigConnections {
		if cc.ComponentID == target.ComponentID {
			tfConfig = cc
			break
		}
	}

	info := &WorkspaceInfo{
		WorkspaceID: target.TerraformWorkspace.ID,
		Name:        componentName,
	}

	if tfConfig != nil && tfConfig.TerraformModule != nil {
		info.TerraformVersion = tfConfig.TerraformModule.Version
		info.VCS = extractComponentVCS(tfConfig.TerraformModule)
	}

	if info.TerraformVersion == "" {
		return nil, fmt.Errorf("unable to determine terraform version for component %q", componentName)
	}

	return info, nil
}

func ListTerraformComponents(ctx context.Context, api nuon.Client, installID string) ([]string, error) {
	components, _, err := api.GetInstallComponents(ctx, installID, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get install components: %w", err)
	}

	var names []string
	for _, c := range components {
		if c.Component != nil && c.Component.Type == "terraform_module" {
			names = append(names, c.Component.Name)
		}
	}
	return names, nil
}

func extractSandboxVCS(sandbox *models.AppAppSandboxConfig) *VCSSource {
	if sandbox.PublicGitVcsConfig != nil {
		return &VCSSource{
			Repo:      sandbox.PublicGitVcsConfig.Repo,
			Directory: sandbox.PublicGitVcsConfig.Directory,
			Branch:    sandbox.PublicGitVcsConfig.Branch,
			IsPublic:  true,
		}
	}
	if sandbox.ConnectedGithubVcsConfig != nil {
		return &VCSSource{
			Repo:      sandbox.ConnectedGithubVcsConfig.Repo,
			Directory: sandbox.ConnectedGithubVcsConfig.Directory,
			Branch:    sandbox.ConnectedGithubVcsConfig.Branch,
			IsPublic:  false,
		}
	}
	return nil
}

func extractComponentVCS(tf *models.AppTerraformModuleComponentConfig) *VCSSource {
	if tf.PublicGitVcsConfig != nil {
		return &VCSSource{
			Repo:      tf.PublicGitVcsConfig.Repo,
			Directory: tf.PublicGitVcsConfig.Directory,
			Branch:    tf.PublicGitVcsConfig.Branch,
			IsPublic:  true,
		}
	}
	if tf.ConnectedGithubVcsConfig != nil {
		return &VCSSource{
			Repo:      tf.ConnectedGithubVcsConfig.Repo,
			Directory: tf.ConnectedGithubVcsConfig.Directory,
			Branch:    tf.ConnectedGithubVcsConfig.Branch,
			IsPublic:  false,
		}
	}
	return nil
}
