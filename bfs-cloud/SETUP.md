# Azure Infrastructure Setup Guide

This guide covers the prerequisites and setup steps for deploying the BFS infrastructure to Azure.

## Prerequisites

### 1. Azure Subscription Requirements

Your Azure subscription must support:
- ✅ Standard resources (VNet, Storage, Cosmos DB, etc.)
- ❌ Cost Management API (requires EA, Web Direct, or MCA offer types)
  - **Note:** Azure Sponsorship subscriptions (MS-AZR-0036P) do not support budget APIs

### 2. Required Resource Providers

Before running Terraform, you must register the following Azure resource providers:

```bash
# Register Microsoft.App for Container Apps
az provider register --namespace Microsoft.App

# Verify registration status (this may take a few minutes)
az provider show --namespace Microsoft.App --query "registrationState"

# Register other required providers
az provider register --namespace Microsoft.Network
az provider register --namespace Microsoft.DocumentDB
az provider register --namespace Microsoft.KeyVault
az provider register --namespace Microsoft.OperationalInsights
az provider register --namespace Microsoft.Insights
```

**Important:** The `Microsoft.App` provider registration can take 5-10 minutes. You must wait until it shows `"Registered"` before running Terraform apply.

### 3. Service Principal Permissions

The Terraform service principal requires **only** the **Contributor** role at the subscription or resource group level.

**To assign the role:**

```bash
# Get your service principal's object ID
SP_OBJECT_ID=$(az ad sp show --id <your-client-id> --query id -o tsv)

# Assign Contributor role at subscription level
az role assignment create \
  --assignee-object-id $SP_OBJECT_ID \
  --role "Contributor" \
  --scope "/subscriptions/<subscription-id>"

# Or assign at resource group level (after creating the RG)
az role assignment create \
  --assignee-object-id $SP_OBJECT_ID \
  --role "Contributor" \
  --scope "/subscriptions/<subscription-id>/resourceGroups/bfs-staging-rg"
```

**Note:** You do **NOT** need "User Access Administrator" role. The infrastructure has been configured to use Key Vault Access Policies instead of RBAC, which eliminates the need for role assignment permissions.

## Known Issues and Solutions

### Issue 1: Microsoft.App Provider Not Registered

**Error:**
```
Error: creating Managed Environment: unexpected status 409 (409 Conflict) with error:
MissingSubscriptionRegistration: The subscription is not registered to use namespace 'Microsoft.App'
```

**Solution:**
```bash
az provider register --namespace Microsoft.App
# Wait 5-10 minutes, then verify:
az provider show --namespace Microsoft.App --query "registrationState"
```

### Issue 2: Cosmos DB Already Exists

**Error:**
```
Error: CosmosDB Account bfs-staging-cosmos already exists, please import the resource
```

**Solution:**
The `staging/main.tf` file already includes an import block. On first run, Terraform will import the existing Cosmos DB instead of trying to create it.

If you need to manually import:
```bash
cd envs/staging
terraform import 'module.bfs_infrastructure.module.cosmos.azurerm_cosmosdb_account.this' \
  '/subscriptions/<sub-id>/resourceGroups/bfs-staging-rg/providers/Microsoft.DocumentDB/databaseAccounts/bfs-staging-cosmos'
```

### Issue 3: Budget API Not Supported

**Error:**
```
Error: Cost Management supports only Enterprise Agreement, Web direct and Microsoft Customer Agreement offer types.
```

**Solution:**
Budget resources have been removed from the Terraform configuration. Azure Sponsorship subscriptions do not support the Cost Management API. You can monitor costs manually in the Azure Portal under Cost Management.

## Deployment Steps

### Staging Environment

```bash
cd bfs-cloud/envs/staging

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the infrastructure
terraform apply
```

### Production Environment

```bash
cd bfs-cloud/envs/production

# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Apply the infrastructure
terraform apply
```

## Infrastructure Components

### Created Resources (Per Environment)

- **Resource Group**: `bfs-{env}-rg`
- **Virtual Network**: `bfs-{env}-vnet` with subnets
  - Container Apps subnet (10.1.0.0/21)
  - Private Endpoints subnet (10.1.8.0/24)
- **Network Security Group**: Allow HTTP/HTTPS and Container Apps traffic
- **Cosmos DB**: MongoDB API database (`bfs-{env}-cosmos`)
- **Key Vault**: Secrets storage with Access Policies (`bfs-{env}-kv`)
- **Container Apps Environment**: `bfs-{env}-env`
- **Container Apps**:
  - `frontend-{env}` (Next.js web app)
  - `backend-{env}` (Go API)
- **User-Assigned Managed Identity**: For Container Apps to access Key Vault
- **Log Analytics Workspace**: Monitoring and diagnostics
- **Action Groups**: Email alerts

### Removed Features

The following features have been removed to work with Contributor-only permissions and Azure Sponsorship subscriptions:

- ❌ RBAC role assignments (replaced with Key Vault Access Policies)
- ❌ Budget alerts (not supported on sponsorship subscriptions)
- ❌ Custom 5xx error metric alerts (removed per user request)

## Authentication Architecture

### Key Vault Access

The infrastructure uses **Access Policies** instead of RBAC:

1. **Terraform Service Principal**: Full admin access (create/read/update/delete secrets)
2. **Container Apps Managed Identity**: Read-only access (get/list secrets)

This approach works with Contributor-only permissions and doesn't require User Access Administrator role.

### How It Works

1. Terraform creates a User-Assigned Managed Identity
2. Container Apps are assigned this identity
3. Key Vault Access Policy grants the identity permission to read secrets
4. Container Apps reference secrets via Key Vault secret URIs
5. At runtime, Azure handles authentication automatically using the managed identity

## Troubleshooting

### Provider Registration Check

```bash
# List all registered providers
az provider list --query "[?registrationState=='Registered'].namespace" -o table

# Check specific provider
az provider show --namespace Microsoft.App
```

### View Terraform State

```bash
# List all resources in state
terraform state list

# Show details of specific resource
terraform state show 'module.bfs_infrastructure.module.cosmos.azurerm_cosmosdb_account.this'
```

### Force Resource Recreation

```bash
# Taint a resource to force recreation on next apply
terraform taint 'module.bfs_infrastructure.module.aca_env.azurerm_container_app_environment.this'
```

## Cost Estimation

**Monthly costs (approximate):**

- Cosmos DB (400 RU/s): ~$24/month
- Container Apps (with scale-to-zero): ~$5-15/month
- Log Analytics: ~$3/month
- Key Vault: ~$1/month
- Networking: Free tier
- **Total**: ~$33-43/month per environment

**Note:** Actual costs vary based on usage. Container Apps scale to zero when idle, significantly reducing costs.
