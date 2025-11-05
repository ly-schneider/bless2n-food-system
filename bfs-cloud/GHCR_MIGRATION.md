# Switch to GitHub Container Registry (GHCR)

This guide shows how to switch from Azure Container Registry to GitHub Container Registry to save CHF 4-6/month.

## Why GHCR?

- **Free** for public and private repositories
- **No storage limits** for reasonable usage
- **Built-in security scanning**
- **No Azure costs**
- Works seamlessly with GitHub Actions

## Prerequisites

- GitHub Personal Access Token (PAT) with `write:packages` permission
- Access to Terraform Cloud or local Terraform

## Step 1: Update Build Workflows

Update `.github/workflows/10-build-push-images.yml` to push to GHCR:

### Current (ACR):
```yaml
- name: Login to Azure Container Registry
  uses: azure/docker-login@v1
  with:
    login-server: ${{ secrets.ACR_NAME }}.azurecr.io
    username: ${{ secrets.AZURE_CLIENT_ID }}
    password: ${{ secrets.AZURE_CLIENT_SECRET }}

- name: Build and push images
  run: |
    docker buildx build --platform linux/amd64 \
      -t ${{ secrets.ACR_NAME }}.azurecr.io/frontend:${{ github.ref_name }} \
      --push bfs-web-app/
```

### New (GHCR):
```yaml
- name: Login to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- name: Build and push images
  run: |
    docker buildx build --platform linux/amd64 \
      -t ghcr.io/${{ github.repository }}/frontend:${{ github.ref_name }} \
      --push bfs-web-app/
```

The `GITHUB_TOKEN` is automatically provided by GitHub Actions - no secret setup needed!

## Step 2: Create GitHub PAT for Container Apps

Container Apps need credentials to pull from GHCR:

1. Go to https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Name: "Container Apps Pull"
4. Scopes: Select `read:packages`
5. Click "Generate token"
6. **Copy the token** (you won't see it again!)

## Step 3: Add PAT to Azure Key Vault

Add the GHCR token as a secret in Key Vault:

```bash
# For production
az keyvault secret set \
  --vault-name bfs-production-kv \
  --name ghcr-token \
  --value "ghp_your_token_here"

# For staging
az keyvault secret set \
  --vault-name bfs-staging-kv \
  --name ghcr-token \
  --value "ghp_your_token_here"
```

## Step 4: Update Terraform Configuration

### Production: `bfs-cloud/envs/production/main.tf`

```hcl
locals {
  registry_host = "ghcr.io/${{ github.repository }}"  # Changed from ACR
  frontend_repo = "frontend"
  backend_repo  = "backend"

  frontend_image = var.frontend_digest != "" ? 
    "${local.registry_host}/${local.frontend_repo}@${var.frontend_digest}" : 
    "${local.registry_host}/${local.frontend_repo}:${var.image_tag}"

  backend_image = var.backend_digest != "" ? 
    "${local.registry_host}/${local.backend_repo}@${var.backend_digest}" : 
    "${local.registry_host}/${local.backend_repo}:${var.image_tag}"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment  = "production"
  location     = var.location
  tags         = var.tags
  alert_emails = var.alert_emails

  config = {
    # ... other settings ...
    
    enable_acr = false  # CHANGED: Disable ACR creation
    # Remove acr_name line
    
    apps = {
      frontend-production = {
        # ... other settings ...
        
        registries = [  # CHANGED: Add GHCR registry
          {
            server               = "ghcr.io"
            username             = "ly-schneider"  # Your GitHub username
            password_secret_name = "ghcr-token"
          }
        ]
        
        # Add GHCR token to Key Vault secrets
        key_vault_secrets = merge(
          lookup(var.app_secrets, "frontend-production", {}),
          {
            "NEXT_PUBLIC_API_BASE_URL" = "next-public-api-base-url"
            "BACKEND_INTERNAL_URL"     = "backend-internal-url"
            "ghcr-token"               = "ghcr-token"  # ADDED
          }
        )
      }
      
      backend-production = {
        # ... other settings ...
        
        registries = [  # CHANGED: Add GHCR registry
          {
            server               = "ghcr.io"
            username             = "ly-schneider"  # Your GitHub username
            password_secret_name = "ghcr-token"
          }
        ]
        
        # Add GHCR token to Key Vault secrets
        key_vault_secrets = merge(
          lookup(var.app_secrets, "backend-production", {}),
          {
            "MONGO_URI"      = "mongo-uri"
            # ... other secrets ...
            "ghcr-token"     = "ghcr-token"  # ADDED
          }
        )
      }
    }
  }
}
```

Do the same for `bfs-cloud/envs/staging/main.tf`.

## Step 5: Deploy Changes

1. **Commit and push changes** to your branch
2. **Merge to production** branch
3. Terraform will:
   - Skip ACR creation (`enable_acr = false`)
   - Configure apps to pull from GHCR
   - Use GitHub PAT from Key Vault

## Step 6: Verify Deployment

Check that apps are running with GHCR images:

```bash
# Check container app status
az containerapp show \
  --name frontend-production \
  --resource-group bfs-production-rg \
  --query "properties.template.containers[0].image"

# Should show: ghcr.io/ly-schneider/bless2n-food-system/frontend:production
```

## Step 7: Delete ACR (After Verification)

Once everything is working with GHCR for **at least a week**:

```bash
# Delete production ACR
az acr delete \
  --name bfsproductionacr \
  --resource-group bfs-production-rg \
  --yes

# Delete staging ACR
az acr delete \
  --name bfsstagingacr \
  --resource-group bfs-staging-rg \
  --yes
```

Or delete via Terraform by removing the ACR module completely.

## Troubleshooting

### Issue: "Failed to pull image"

**Solution:** Check that:
1. GHCR token is correctly stored in Key Vault
2. Token has `read:packages` permission
3. Container image exists at `ghcr.io/ly-schneider/bless2n-food-system/frontend`

### Issue: "Image not found"

**Solution:** Verify image was pushed successfully:

```bash
# List packages in your repository
gh api /user/packages/container
```

Or check at https://github.com/ly-schneider?tab=packages

### Issue: "Authentication failed"

**Solution:** 
1. Regenerate GitHub PAT
2. Update in Key Vault
3. Restart Container App

## Rollback Plan

If you need to rollback to ACR:

1. Revert Terraform changes
2. Set `enable_acr = true`
3. Remove `registries` blocks
4. Deploy

Your ACR images should still exist (unless you deleted the ACR).

## Cost Savings

| Item | Before | After | Savings |
|------|--------|-------|---------|
| Production ACR | CHF 4-6/month | CHF 0 | CHF 4-6 |
| Staging ACR | CHF 4-6/month | CHF 0 | CHF 4-6 |
| GHCR | CHF 0 | CHF 0 | - |
| **Total** | **CHF 8-12/month** | **CHF 0** | **CHF 8-12** |

Combined with other optimizations, this helps reach the CHF 10/month target!
