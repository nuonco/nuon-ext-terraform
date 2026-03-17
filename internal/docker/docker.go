package docker

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const baseImage = "hashicorp/terraform"

type RunConfig struct {
	TerraformVersion string
	WorkspaceID      string
	OrgID            string
	APIURL           string
	APIToken         string
	SourceDir        string
	Command          []string
}

func EnsureImage(tfVersion string) error {
	image := baseImage + ":" + tfVersion
	check := exec.Command("docker", "image", "inspect", image)
	check.Stdout = nil
	check.Stderr = nil
	if check.Run() == nil {
		return nil
	}

	fmt.Fprintf(os.Stderr, "Pulling %s...\n", image)
	pull := exec.Command("docker", "pull", image)
	pull.Stdout = os.Stderr
	pull.Stderr = os.Stderr
	return pull.Run()
}

func Run(cfg RunConfig) error {
	image := baseImage + ":" + cfg.TerraformVersion

	sourceDir, err := filepath.Abs(cfg.SourceDir)
	if err != nil {
		return fmt.Errorf("unable to resolve source directory: %w", err)
	}

	if _, err := os.Stat(sourceDir); err != nil {
		return fmt.Errorf("source directory does not exist: %s", sourceDir)
	}

	if len(cfg.Command) > 0 && cfg.Command[0] == "shell" {
		args := []string{
			"run", "--rm", "-it",
			"--entrypoint", "/bin/sh",
			"-v", sourceDir + ":/workspace",
			"-w", "/workspace",
			"-e", "TF_HTTP_AUTHORIZATION=Bearer " + cfg.APIToken,
			image,
		}
		return dockerExec(args)
	}

	args := []string{
		"run", "--rm", "-i",
		"-v", sourceDir + ":/workspace",
		"-w", "/workspace",
		"-e", "TF_HTTP_AUTHORIZATION=Bearer " + cfg.APIToken,
		image,
	}
	args = append(args, cfg.Command...)
	return dockerExec(args)
}

func RunInit(cfg RunConfig) error {
	image := baseImage + ":" + cfg.TerraformVersion

	sourceDir, err := filepath.Abs(cfg.SourceDir)
	if err != nil {
		return fmt.Errorf("unable to resolve source directory: %w", err)
	}

	backendContent := generateBackendConfig(cfg.APIURL, cfg.WorkspaceID, cfg.OrgID)

	tmpDir, err := os.MkdirTemp("", "nuon-tf-backend-*")
	if err != nil {
		return fmt.Errorf("unable to create temp dir for backend config: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	backendFile := filepath.Join(tmpDir, "nuon_backend.tf")
	if err := os.WriteFile(backendFile, []byte(backendContent), 0o644); err != nil {
		return fmt.Errorf("unable to write backend config: %w", err)
	}

	args := []string{
		"run", "--rm", "-i",
		"--entrypoint", "/bin/sh",
		"-v", sourceDir + ":/workspace",
		"-v", tmpDir + ":/nuon-backend:ro",
		"-w", "/workspace",
		"-e", "TF_HTTP_AUTHORIZATION=Bearer " + cfg.APIToken,
		image,
		"-c", "cp /nuon-backend/nuon_backend.tf /workspace/nuon_backend.tf && terraform init -reconfigure",
	}
	return dockerExec(args)
}

func dockerExec(args []string) error {
	cmd := exec.Command("docker", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func generateBackendConfig(apiURL, workspaceID, orgID string) string {
	return fmt.Sprintf(`terraform {
  backend "http" {
    lock_method    = "POST"
    unlock_method  = "POST"
    address        = "%s/v1/terraform-backend?workspace_id=%s&org_id=%s"
    lock_address   = "%s/v1/terraform-workspaces/%s/lock?org_id=%s"
    unlock_address = "%s/v1/terraform-workspaces/%s/unlock?org_id=%s"
  }
}
`, apiURL, workspaceID, orgID,
		apiURL, workspaceID, orgID,
		apiURL, workspaceID, orgID)
}

// CloneResult holds the paths from a public repo clone.
type CloneResult struct {
	// SourceDir is the directory containing the Terraform files (may be a
	// subdirectory of CloneRoot).
	SourceDir string
	// CloneRoot is the top-level temp directory to remove on cleanup.
	CloneRoot string
}

func ClonePublicRepo(repo, branch, directory string) (*CloneResult, error) {
	repoURL := repo
	if !strings.Contains(repoURL, "://") {
		repoURL = "https://github.com/" + repo + ".git"
	}

	tmpDir, err := os.MkdirTemp("", "nuon-tf-source-*")
	if err != nil {
		return nil, fmt.Errorf("unable to create temp dir: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Cloning %s (branch: %s)...\n", repo, branch)
	cmd := exec.Command("git", "clone", "--depth", "1", "--branch", branch, repoURL, tmpDir)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("unable to clone %s: %w", repo, err)
	}

	sourceDir := tmpDir
	if directory != "" && directory != "." {
		sourceDir = filepath.Join(tmpDir, directory)
	}

	if _, err := os.Stat(sourceDir); err != nil {
		os.RemoveAll(tmpDir)
		return nil, fmt.Errorf("directory %q not found in repo", directory)
	}

	return &CloneResult{
		SourceDir: sourceDir,
		CloneRoot: tmpDir,
	}, nil
}
