# nuon-ext-terraform

CLI extension that drops you into an interactive Terraform shell connected to a Nuon-managed sandbox or component workspace.

## What it does

1. Resolves the Terraform workspace ID, version, and source code from the Nuon API
2. Spins up a Docker container with the exact `hashicorp/terraform` version configured for the workspace
3. Generates and injects the `nuon_backend.tf` HTTP backend config with auth
4. Runs `terraform init -reconfigure` automatically
5. Drops you into an interactive shell where you can run Terraform commands directly
6. For public repos, clones the source automatically; for connected repos, you provide the local path

## Prerequisites

- **Docker** -- must be installed and running
- **nuon CLI** -- authenticated with an org selected (`nuon login` + `nuon orgs select`)
- **Install selected** -- via `nuon installs select`, `NUON_INSTALL_ID`, or `--install-id` flag

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
# Drop into a Terraform shell for the sandbox workspace
nuon terraform sandbox

# Use a local source directory (required for connected repos)
nuon terraform sandbox --source-dir ~/code/my-sandbox

# Override the install ID
nuon terraform sandbox --install-id ins123
```

### Component

```bash
# Drop into a Terraform shell for a specific component
nuon terraform component --name certificate_wildcard_public

# With a local source directory
nuon terraform component --name my-database --source-dir ~/code/my-infra
```

Once inside the shell, run Terraform commands directly:

```
/workspace $ terraform plan
/workspace $ terraform state list
/workspace $ terraform state show aws_eks_cluster.main
/workspace $ terraform apply
```

### Flags

| Flag | Description |
|------|-------------|
| `--install-id` | Override the install ID (defaults to selected install or `NUON_INSTALL_ID`) |
| `--name` | Component name (required for `component` subcommand) |
| `--source-dir` | Local directory with Terraform source (required for connected repos) |

### Install ID resolution

The extension resolves the install ID in this order:

1. `--install-id` flag
2. `NUON_INSTALL_ID` environment variable
3. `install_id` from the CLI config file (`~/.nuon`), set by `nuon installs select`

### Connected repos

When the sandbox or component uses a connected repo (private GitHub repo), the extension cannot clone it automatically. Provide the local checkout path:

```bash
nuon terraform sandbox --source-dir ~/code/my-sandbox
nuon terraform component --name my-db --source-dir ~/code/my-infra/modules/database
```

### Custom config file

The extension respects the `nuon -f` flag for using a config file other than `~/.nuon`:

```bash
nuon -f ~/.nuon-staging terraform sandbox
```

## How it works

- The extension uses the nuon-go SDK to call `GetInstall`, `GetInstallComponents`, and `GetAppConfig` to resolve workspace IDs, Terraform versions, and VCS source info.
- It generates a `nuon_backend.tf` file pointing to the ctl-api HTTP backend endpoints (`/v1/terraform-backend` for state, `/v1/terraform-workspaces/{id}/lock` and `/unlock` for locking), with the API token passed as a query parameter.
- The Docker container uses the `hashicorp/terraform:<version>` image, mounts the source at `/workspace`, copies in the backend config, runs `terraform init -reconfigure`, and then drops into `/bin/sh`.

## Development

```bash
# Build (requires NUON_REPO_ROOT pointing to the monorepo)
make build

# Run directly
./nuon-ext-terraform sandbox --help
```
