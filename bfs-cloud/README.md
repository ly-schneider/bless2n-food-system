# BFS Cloud

Terraform infrastructure-as-code for the Bless2n Food System — deploys to Azure with Container Apps, auto-scaling, scale-to-zero, and environment isolation.

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
│  │   ACR    │  │ Key Vault │  │ Blob Storage │      │
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
- **Three-phase CI/CD** — ACR infrastructure first, then image build, then Container Apps deployment
- **Least-privilege RBAC** — minimal Azure roles for the Terraform Cloud service principal

## Module Structure

```
bfs-cloud/
├── envs/
│   ├── staging/                Staging root module
│   │   ├── main.tf             Calls modules/stack with staging config
│   │   ├── backend.tf          State backend (Terraform Cloud)
│   │   ├── variables.tf        Staging-specific variables
│   │   └── outputs.tf
│   └── production/             Production root module
│       ├── main.tf
│       ├── backend.tf
│       ├── variables.tf
│       └── outputs.tf
├── modules/
│   ├── stack/                  Full infrastructure composition
│   ├── containerapp/           Individual container app definition
│   ├── containerapps_env/      Container Apps environment + VNet
│   ├── observability/          Log Analytics workspace
│   ├── diagnostic_setting/     Monitoring & diagnostics
│   ├── alerts/                 Metric-based alerts
│   ├── network/                VNets & subnets
│   ├── blob_storage/           Azure Blob Storage
│   ├── rg/                     Resource group
│   ├── security/               Key Vault & RBAC
│   └── rbac_tfc/               Terraform Cloud service principal roles
├── defaults/                   Default variable values
├── SETUP.md                    Prerequisites & troubleshooting
└── README.md
```

## Auto-Scaling Configuration

Both environments support multi-signal scaling:

| Environment | App | Replicas | HTTP Threshold | CPU | Memory |
|-------------|-----|----------|---------------|-----|--------|
| **Production** | Frontend | 0–20 | 30 concurrent | 70% | — |
| **Production** | Backend | 0–20 | 50 concurrent | 80% | 85% |
| **Staging** | Frontend | 0–20 | 20 concurrent | 75% | — |
| **Staging** | Backend | 0–20 | 40 concurrent | 80% | — |

## Usage

### Deploy

```bash
cd envs/staging
terraform init
terraform plan
terraform apply
```

### CI/CD Three-Phase Deployment

1. **Phase 1 — ACR Infrastructure**: Terraform creates Azure Container Registry
2. **Phase 2 — Image Build**: GitHub Actions builds and pushes Docker images to ACR
3. **Phase 3 — Container Apps**: Terraform deploys apps referencing the new images

### Terraform Variables

| Variable | Type | Purpose |
|----------|------|---------|
| `enable_acr` | bool | Create ACR and grant AcrPull to identities |
| `acr_name` | string | ACR name (images resolve to `<name>.azurecr.io`) |
| `image_tag` | string | Branch/version tag to deploy |
| `frontend_digest` / `backend_digest` | string | Immutable image digest (preferred in CI) |
| `revision_suffix` | string | Force new revision (e.g., commit SHA) |

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

### RBAC for Terraform Cloud

The `modules/rbac_tfc` module grants minimal roles at the resource group scope:

- **Network Contributor** — VNets, subnets, private endpoints
- **Private DNS Zone Contributor** — DNS zones/links
- **User Access Administrator** — RBAC assignments (Key Vault, ACR Pull)
- **Managed Identity Contributor** — user-assigned identities for Container Apps

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
