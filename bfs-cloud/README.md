# Bless2n Food System (BFS) — Azure Terraform

## Architecture Overview

This Terraform setup uses a single, concrete approach: deploy from `envs/<env>` (staging, production) as the root module. Shared building blocks live under `modules/`. No separate `common/` or `env-configs/` flow.

## What this deploys

### Infrastructure Components
- **Per-environment isolation**: Resource Groups, VNets, delegated Subnets, Container Apps Environment
- **Container Apps**: Auto-scaling web applications with intelligent scaling rules
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

### Two-State Layout (Staging & Prod)
- Use `envs/staging` and `envs/production` to maintain separate Terraform states and configurations.
- Each env has its own `backend.tf` (local by default); switch to a remote backend for team usage.

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

### CI/CD Deployment (GitHub Actions + GHCR)

Images are sourced from GHCR and the tag is injected via `image_tag`.

Key variables for CI:
- `registry_server` (default `ghcr.io`)
- `registry_namespace` (e.g., your GitHub org/user)
- `registry_username` (e.g., `${{ github.actor }}`)
- `registry_token` (GHCR token; mark as secret)
- `image_tag` (e.g., `${{ github.sha }}` or a release tag)

Example Action step (assuming Azure credentials already set via OIDC/secrets):

```yaml
    - name: Terraform Apply (staging)
      working-directory: bfs-cloud/envs/staging
      env:
        TF_VAR_registry_server: ghcr.io
        TF_VAR_registry_namespace: ${{ github.repository_owner }}
        TF_VAR_registry_username: ${{ github.actor }}
        TF_VAR_registry_token: ${{ secrets.GHCR_TOKEN }}
        TF_VAR_image_tag: ${{ github.sha }}
      run: |
        terraform init -input=false
        terraform apply -auto-approve -input=false
```

With this, Container Apps pull private images from GHCR using the provided credentials. The same secret is applied to all apps unless overridden per app.

### Providing App Secrets and Registries

Provide secrets and registries via env variables or tfvars at the env root:
- Use `TF_VAR_registry_*` and `TF_VAR_image_tag` to handle GHCR in CI.
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
