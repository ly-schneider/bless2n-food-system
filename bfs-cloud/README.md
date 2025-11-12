# Bless2n Food System (BFS) — Azure Terraform

## Architecture Overview

This Terraform setup uses a single, concrete approach: deploy from `envs/<env>` (staging, production) as the root module. Shared building blocks live under `modules/`. No separate `common/` or `env-configs/` flow.

## What this deploys

### Infrastructure Components
- **Per-environment isolation**: Resource Groups, VNets, delegated Subnets, Container Apps Environment
- **Container Apps**: Auto-scaling web applications with intelligent scaling rules
  - HTTPS ingress enabled by default (no standalone IPs)
  - Default `azurecontainerapps.io` FQDNs for each app
  - Optional custom domains with ACA-managed certificates and DNS CNAMEs
- **Database**: Cosmos DB with MongoDB API
- **Observability**: Log Analytics (both envs), Diagnostic Settings
- **Monitoring**: Metric-based alerts and monitoring (configurable per environment)

### Auto-Scaling Configuration
- **Scale-to-Zero**: Applications scale down to 0 replicas when idle (cost-efficient)
- **Burst Scaling**: Can scale up to 20 replicas during traffic spikes
- **Smart Triggers**: HTTP request-based, CPU percentage, and memory percentage scaling
- **Environment-Specific**: Scaling thresholds for staging

## Project Structure

```
bfs-cloud/
├── envs/
│   ├── staging/             # Staging root module
│   │   ├── backend.tf
│   │   ├── main.tf          # Calls modules/stack with env config
│   │   ├── variables.tf     # Staging variables for CI/manual
│   │   └── outputs.tf
│   └── production/                # Prod root module
│       ├── backend.tf
│       ├── main.tf          # Calls modules/stack with env config
│       ├── variables.tf
│       └── outputs.tf
└── modules/                 # Reusable Terraform modules
    ├── stack/               # Full infra composition (was common/)
    ├── containerapp/
    ├── containerapps_env/
    ├── observability/
    ├── diagnostic_setting/
    ├── alerts/
    ├── cosmos_mongo/
    ├── network/
    ├── rg/
    └── security/
```

## Application Scaling Strategy

### Current Configuration

**Production Apps:**
- **Frontend**: Scales 0-20 replicas based on 30 concurrent requests + 70% CPU
- **Backend**: Scales 0-20 replicas based on 50 concurrent requests + 80% CPU + 85% memory

**Staging Apps:**
- **Frontend**: Scales 0-20 replicas based on 20 concurrent requests + 75% CPU  
- **Backend**: Scales 0-20 replicas based on 40 concurrent requests + 80% CPU

### Scaling Benefits
- **Cost Optimization**: Zero cost when applications are idle
- **Performance**: Automatic scaling based on real demand
- **Reliability**: Multiple scaling triggers prevent bottlenecks
- **Environment-Appropriate**: Thresholds for staging

## Usage

### Standard Deployment
```bash
# Deploy staging  
cd envs/staging
terraform init
terraform apply
```

### App URLs via Key Vault
- Terraform provisions a Key Vault; apps reference the following secret names (not auto-created to avoid overwriting existing values):
  - `backend-url`: public HTTPS URL for the backend (used by frontend `NEXT_PUBLIC_API_BASE_URL` and `BACKEND_INTERNAL_URL`).
  - `frontend-url`: public HTTPS URL for the frontend (used by backend `SECURITY_TRUSTED_ORIGINS`, `JWT_ISSUER`, `PUBLIC_BASE_URL`).
- Create or update these secrets in Azure Portal > Key Vaults > `<kv-name>` > Secrets, or via CI/CD.
- The previous `shared_config` module and its separate env have been removed; no extra Terraform step is required.

### Custom Domains (Managed Certs + Delegated Subzone)
- Each app supports an optional `custom_domains` list to bind hostnames using ACA-managed certificates.
- Production creates an Azure DNS zone for `food.bless2n.ch` only. The parent zone `bless2n.ch` stays at your provider.
- With the current config, Terraform does not create TXT/CNAME records; you will add them manually after delegating the subdomain.
- Apex records are not supported with CNAMEs; use subdomains under `food.bless2n.ch`.

Example per-app config under `config.apps` in `envs/<env>/main.tf`:
```hcl
apps = {
  frontend-staging = {
    # ...existing config...
    custom_domains = [
      {
        hostname                     = "app.example.com"
        dns_zone_name                = "example.com"
        dns_zone_resource_group_name = "dns-zones-rg" # existing Azure DNS zone RG
        ttl                          = 300             # optional
      }
    ]
  }
}
```

What Terraform creates now:
- `azurerm_dns_zone`: `food.bless2n.ch` (in production only)
- `azurerm_container_app_custom_domain` can be enabled later (once DNS is in place)

Manual steps (current configuration):
- Delegate `food.bless2n.ch` at your existing DNS provider by adding NS records pointing to the nameservers output from Terraform.
- Add records at the delegated Azure DNS zone (or at your provider if you don’t delegate):
  - TXT `asuid.food.bless2n.ch` = `<frontend custom_domain_verification_id>`
  - TXT `asuid.api.food.bless2n.ch` = `<backend custom_domain_verification_id>`
  - TXT `asuid.staging.food.bless2n.ch` = `<staging FE verification_id>`
  - TXT `asuid.api.staging.food.bless2n.ch` = `<staging API verification_id>`
  - CNAME `food.bless2n.ch` → `<frontend app azurecontainerapps.io fqdn>`
  - CNAME `api.food.bless2n.ch` → `<backend app azurecontainerapps.io fqdn>`
  - CNAME `staging.food.bless2n.ch` → `<staging frontend fqdn>`
  - CNAME `api.staging.food.bless2n.ch` → `<staging backend fqdn>`

Once DNS is set and propagated, run `terraform apply` again so the `azurerm_container_app_custom_domain` resources can succeed and ACA-managed certificates can be issued and bound. If DNS is not ready yet, the apply may fail during domain binding; create DNS first, then re-apply.

### Two-State Layout (Staging & Prod)
- Use `envs/staging` and `envs/production` to maintain separate Terraform states and configurations.
- Each env has its own `backend.tf` (local by default); switch to a remote backend for team usage.

Remote backend example (Terraform Cloud — recommended):
```hcl
terraform {
  cloud {
    organization = "leys-services"           # replace with your org
    workspaces {
      name = "bfs-staging"              # or bfs-production under envs/prod
    }
  }
}
```

Remote backend example (AzureRM):
```hcl
terraform {
  backend "azurerm" {
    resource_group_name  = "bfs-tfstate-rg"
    storage_account_name = "bfstfstorproduction"
    container_name       = "tfstate"
    key                  = "envs/staging/terraform.tfstate"
  }
}
```
Repeat with `key = "envs/production/terraform.tfstate"` for production.

### CI/CD Deployment: Three-Phase Deployment with ACR

The deployment workflow is split into three phases to ensure proper dependency order:

1. **Deploy ACR Infrastructure**: Creates the Azure Container Registry first
2. **Build & Push Images**: Builds and pushes images to the newly created ACR
3. **Deploy Container Apps**: Deploys the application infrastructure that references the images

**Deployment Flow:**
- **ACR Creation**: Terraform creates ACR using the `modules/acr` module on first run
- **Image Build**: GitHub Actions builds and pushes images to ACR after ACR exists
- **App Deployment**: Container Apps are deployed with references to ACR images

**Terraform Variables per Environment:**
- `enable_acr` (bool): when true, creates ACR and grants `AcrPull` to Container Apps identities
- `acr_name` (string): ACR name; images resolve to `<acr_name>.azurecr.io/<repo>:<tag>`
- `image_tag` (string): branch tag to deploy (e.g., `staging`, `production`).
- `frontend_digest` / `backend_digest` (string, optional): when provided, deployment references images by immutable digest (`<repo>@sha256:...`) instead of mutable tags. Prefer digest-first in CI to guarantee exact images roll out (Buildx outputs digests; Actions exposes them as outputs).
- `revision_suffix` (string, optional): unique value to force a new Container Apps revision (e.g., commit SHA). Keeps image references tag-based while ensuring rollout on each build.

**For Development with GHCR:**
If `enable_acr` is false, the env can use GHCR via:
- `registry_server`, `registry_namespace`, `registry_username`, `registry_token`

**Required GitHub Environment Configuration:**
- `ACR_NAME`: the ACR resource name (e.g., "bfsstagingacr", "bfsprodacr")
- Azure OIDC secrets: `AZURE_CLIENT_ID`, `AZURE_TENANT_ID`, `AZURE_SUBSCRIPTION_ID`

**Terraform Cloud Configuration:**
- Workspace variables: `enable_acr=true`, `acr_name=<your-acr-name>`
- Provider credentials: `ARM_CLIENT_ID`, `ARM_CLIENT_SECRET`, `ARM_TENANT_ID`, `ARM_SUBSCRIPTION_ID`

### Least-Privilege RBAC for Terraform Cloud

The stack includes a module (`modules/rbac_tfc`) that grants the minimal roles your Terraform Cloud service principal needs at the environment Resource Group scope:
- Network Contributor: manage VNets/subnets and Private Endpoints
- Private DNS Zone Contributor: manage Private DNS zones/links (when enabled)
- Cosmos DB Account Contributor: create/update Cosmos resources
- User Access Administrator: create RBAC role assignments (Key Vault access, ACR Pull, etc.)
- Managed Identity Contributor: create user-assigned identities used by Container Apps

By default, the module targets the current Terraform principal. If you want to target a different principal (e.g., bootstrap run), pass its client ID into the module via `principal_client_id` and ensure the AzureAD provider is configured.

Verify effective assignments after apply:
```
az role assignment list --assignee-object-id $(az ad sp show --id $AZURE_CLIENT_ID --query id -o tsv) \
  --scope $(az group show -n <rg-name> --query id -o tsv) --include-inherited -o table
```

Cosmos keys vs AAD note:
- If local auth is disabled on Cosmos, keys cannot be listed. Switch apps to AAD data-plane auth and set the RBAC module to avoid key-related permissions.

### Providing App Secrets and Registries

Provide secrets and registries via env variables or tfvars at the env root:
- Use `TF_VAR_registry_*`, `TF_VAR_image_tag` (branch), and optionally `TF_VAR_frontend_digest` / `TF_VAR_backend_digest` (immutable digests) and `TF_VAR_revision_suffix` (commit SHA) in CI.
- Optionally pass per-app overrides via `TF_VAR_app_secrets` and `TF_VAR_app_registries`.
  - Example: `TF_VAR_app_secrets='{"frontend-staging":{"API_KEY":"..."}}'`
These propagate to Azure Container Apps as `secret {}` and `registry {}` blocks.

The previous `common/` + `env-configs/*.tfvars` path has been removed in favor of the simpler env roots.

## Key Features

### Unified State Management
- **Single Source of Truth**: All infrastructure logic in `common/main.tf`
- **Environment Consistency**: Same infrastructure behavior across environments
- **Easy Maintenance**: Changes made once, applied everywhere
- **Configuration-Driven**: Environment differences handled through variables

### Auto-Scaling Capabilities
- **HTTP Scale Rules**: Scale based on concurrent requests
- **CPU Scale Rules**: Scale based on CPU percentage
- **Memory Scale Rules**: Scale based on memory percentage  
- **Azure Queue Scale Rules**: Scale based on queue length (configurable)
- **Custom Scale Rules**: Extensible for custom metrics

### Environment Isolation
- **Network Isolation**: Separate VNets and subnets per environment
- **Resource Isolation**: Dedicated resource groups per environment
- **Configuration Flexibility**: Different scaling, monitoring, and resource settings

## Customization

### Adding New Environments
1. Create new directory under `envs/`
2. Copy configuration from existing environment
3. Modify environment-specific variables
4. Deploy using standard Terraform workflow

### Modifying Scaling Rules
Edit the `apps` configuration in environment main.tf files:

```hcl
apps = {
  frontend = {
    # ... basic config ...
    http_scale_rule = {
      name                = "custom-http-scale"
      concurrent_requests = 25  # Adjust threshold
    }
    cpu_scale_rule = {
      name           = "custom-cpu-scale" 
      cpu_percentage = 75  # Adjust threshold
    }
  }
}
```
