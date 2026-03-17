package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/nuonco/nuon/bins/cli/extensions/nuon-ext-terraform/internal/docker"
	"github.com/nuonco/nuon/bins/cli/extensions/nuon-ext-terraform/internal/resolve"
)

// resolveSource determines the Terraform source directory. For public repos it
// clones automatically. For connected repos it requires --source-dir.
//
// Returns (sourceDir, cleanupDir, error). If cleanupDir is non-empty, the
// caller should defer os.RemoveAll(cleanupDir) to clean up the temp clone.
func resolveSource(info *resolve.WorkspaceInfo, sourceDirFlag string) (string, string, error) {
	if sourceDirFlag != "" {
		abs, err := filepath.Abs(sourceDirFlag)
		if err != nil {
			return "", "", fmt.Errorf("unable to resolve --source-dir: %w", err)
		}
		if _, err := os.Stat(abs); err != nil {
			return "", "", fmt.Errorf("source directory does not exist: %s", abs)
		}
		fmt.Fprintf(os.Stderr, "Source:     %s\n", abs)
		return abs, "", nil
	}

	if info.VCS == nil {
		return "", "", fmt.Errorf("unable to determine source repo; provide --source-dir")
	}

	if info.VCS.IsPublic {
		fmt.Fprintf(os.Stderr, "Source:     %s (branch: %s, dir: %s)\n", info.VCS.Repo, info.VCS.Branch, info.VCS.Directory)
		result, err := docker.ClonePublicRepo(info.VCS.Repo, info.VCS.Branch, info.VCS.Directory)
		if err != nil {
			return "", "", err
		}
		return result.SourceDir, result.CloneRoot, nil
	}

	return "", "", fmt.Errorf(
		"component %q uses a connected repo (%s); provide the local source directory with --source-dir",
		info.Name, info.VCS.Repo,
	)
}
