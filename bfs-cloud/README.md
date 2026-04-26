# BFS Cloud

Terraform infrastructure-as-code for the BlessThun Food System — deploys to Azure with Container Apps, auto-scaling, scale-to-zero, and environment isolation.

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                  Azure Resource Group               │
├─────────────────────────────────────────────────────┤
│                                                     │
│  ┌──────────────────────────────────────────────┐   │
│  │         Container Apps Environment           │   │
│  │              (VNet delegated)                │   │
│  │                                              │   │
│  │  ┌────────────────┐   ┌─────────────────┐    │   │
│  │  │  Frontend      │   │  Backend        │    │   │
│  │  │  0-20 replicas │   │  0-20 replicas  │    │   │
│  │  │  (Next.js)     │   │  (Go)           │    │   │
│  │  └────────────────┘   └─────────────────┘    │   │
│  └──────────────────────────────────────────────┘   │
│                                                     │
│  ┌──────────┐  ┌───────────┐  ┌──────────────┐      │
│  │   GHCR   │  │ Key Vault │  │ Blob Storage │      │
│  └──────────┘  └───────────┘  └──────────────┘      │
│                                                     │
│  ┌──────────────────┐  ┌──────────────────────┐     │
│  │  Log Analytics   │  │  Diagnostic Settings │     │
│  └──────────────────┘  └──────────────────────┘     │
└─────────────────────────────────────────────────────┘
```

### Key Design Decisions

- **Scale-to-zero** — zero cost when applications are idle, auto-burst up to 20 replicas
- **Immutable image promotion** — production uses the exact staging image digest (no rebuild)
- **Environment isolation** — separate VNets, resource groups, and Terraform states per environment
- **GHCR image registry** — images are built and pushed to GitHub Container Registry, pulled into Container Apps via token
- **External database** — PostgreSQL hosted on NeonDB (managed outside Terraform)

## Module Structure

```
bfs-cloud/
├── envs/
│   ├── staging/                Staging root module
│   │   ├── main.tf             Calls modules/stack with staging config
│   │   ├── versions.tf         Provider versions & Terraform Cloud backend
│   │   ├── variables.tf        Staging-specific variables
│   │   ├── imports.tf          Terraform import blocks
│   │   └── outputs.tf
│   └── production/             Production root module
│       ├── main.tf
│       ├── versions.tf
│       ├── variables.tf
│       ├── imports.tf
│       └── outputs.tf
├── modules/
│   ├── stack/                  Full infrastructure composition
│   ├── containerapp/           Individual container app definition
│   ├── containerapps_env/      Container Apps environment + VNet
│   ├── observability/          Log Analytics workspace
│   ├── diagnostic_setting/     Monitoring & diagnostics
│   ├── blob_storage/           Azure Blob Storage
│   ├── rg/                     Resource group
│   └── security/               Key Vault & RBAC
├── defaults/                   Default variable values
├── SETUP.md                    Prerequisites & troubleshooting
└── README.md
```

## Auto-Scaling Configuration

Both environments support multi-signal scaling:

| Environment    | App      | Replicas | HTTP Threshold | CPU | Memory |
| -------------- | -------- | -------- | -------------- | --- | ------ |
| **Production** | Frontend | 0–20     | 30 concurrent  | 70% | —      |
| **Production** | Backend  | 0–20     | 50 concurrent  | 80% | 85%    |
| **Staging**    | Frontend | 0–20     | 20 concurrent  | 75% | —      |
| **Staging**    | Backend  | 0–20     | 40 concurrent  | 80% | —      |

## Usage

### Deploy

```bash
cd envs/staging
terraform init
terraform plan
terraform apply
```

### CI/CD Deployment

1. **Image Build**: GitHub Actions builds and pushes Docker images to GHCR
2. **Container Apps**: Terraform deploys apps pulling images from GHCR via token

### Terraform Variables

| Variable                                             | Type   | Purpose                                  |
| ---------------------------------------------------- | ------ | ---------------------------------------- |
| `ghcr_token`                                         | string | GitHub Container Registry pull token     |
| `image_tag`                                          | string | Branch/version tag to deploy             |
| `frontend_digest` / `backend_digest` / `docs_digest` | string | Immutable image digest (preferred in CI) |
| `revision_suffix`                                    | string | Force new revision (e.g., commit SHA)    |

### App URLs via Key Vault

Terraform provisions a Key Vault with these secret names (values managed externally):

- `backend-url` — public HTTPS URL for the backend
- `frontend-url` — public HTTPS URL for the frontend

### Remote Backend

Terraform Cloud (recommended):

```hcl
terraform {
  cloud {
    organization = "your-org"
    workspaces {
      name = "bfs-staging"    # or bfs-production
    }
  }
}
```

### Adding a New Environment

1. Create `envs/<name>/`
2. Copy from an existing environment
3. Adjust variables (scaling thresholds, resource names)
4. Configure Terraform state backend
5. `terraform init && terraform apply`

### Modifying Scaling Rules

Edit the `apps` block in the environment's `main.tf`:

```hcl
apps = {
  frontend = {
    http_scale_rule = {
      name                = "http-scale"
      concurrent_requests = 25
    }
    cpu_scale_rule = {
      name           = "cpu-scale"
      cpu_percentage = 75
    }
  }
}
```
