# nuon-ext-terraform

CLI extension for running Terraform commands against Nuon-managed sandbox and component workspaces via Docker containers.

## What it does

- Resolves the Terraform workspace ID, version, and source code from the Nuon API
- Spins up a Docker container with the exact `hashicorp/terraform` version configured for the workspace
- Generates and injects the `nuon_backend.tf` HTTP backend config automatically on `init`
- For public repos, clones the source automatically; for connected repos, you provide the local path

## Install

```bash
nuon ext install ./nuon-ext-terraform
```

Or build from the monorepo:

```bash
make build
nuon ext install .
```

## Usage

### Sandbox

```bash
# Initialize the backend (creates nuon_backend.tf and runs terraform init)
nuon terraform sandbox init

# Preview changes
nuon terraform sandbox plan

# Apply changes
nuon terraform sandbox apply

# Inspect state
nuon terraform sandbox state list
nuon terraform sandbox state show aws_eks_cluster.main

# Drop into a shell inside the container
nuon terraform sandbox shell
```

### Component

```bash
# Initialize a component workspace
nuon terraform component init --name certificate_wildcard_public

# Plan
nuon terraform component plan --name my-database

# State operations
nuon terraform component state list --name my-database
nuon terraform component state show aws_rds_cluster.main --name my-database
```

### Flags

| Flag | Description |
|------|-------------|
| `--install-id` | Override the install ID (defaults to `NUON_INSTALL_ID`) |
| `--name` | Component name (required for `component` subcommand) |
| `--source-dir` | Local directory with Terraform source (required for connected repos) |

### Connected repos

When the sandbox or component uses a `connected_repo` (private GitHub repo), the extension cannot clone it automatically. Provide the local checkout path:

```bash
nuon terraform sandbox init --source-dir ~/code/my-sandbox
nuon terraform component plan --name my-db --source-dir ~/code/my-infra/modules/database
```

## Requirements

- Docker
- `nuon` CLI authenticated with an org selected
- An install ID set via `nuon installs select` or `--install-id`

## Development

```bash
# Build (requires NUON_REPO_ROOT pointing to the monorepo)
make build

# Run directly
./nuon-ext-terraform sandbox --help
```
