# Azure Cost Optimization Guide

This document explains how to reduce Azure infrastructure costs to under CHF 10/month.

## Understanding the Cost Problem

Even when Container Apps are scaled to 0 replicas, these resources **still incur daily costs**:

1. **Container Apps Environment** (~CHF 8-12/month)
   - The environment infrastructure itself has a base cost
   - Exists even when all apps are scaled to 0

2. **Virtual Network** (~CHF 2-4/month)
   - Dedicated subnet and network infrastructure
   - Always allocated once created

3. **Log Analytics Workspace** (~CHF 2-5/month)
   - Base operational costs
   - Even with minimal data ingestion

4. **Container Registry** (~CHF 4-6/month for Basic SKU)
   - Storage costs for images
   - Operational costs even when not actively used

**Total estimated cost when scaled to 0: ~CHF 16-27/month**

## Solutions: Reduce Costs Without Deleting Resources

If destroying infrastructure is not an option, you can still reduce costs significantly through optimization:

### Quick Wins (No Resource Deletion)

| Optimization | Savings/Month | Difficulty |
|-------------|---------------|------------|
| Use GitHub Container Registry | CHF 4-6 | Medium |
| Share ACR between environments | CHF 4-6 | Medium |
| Reduce Cosmos DB diagnostic logging | CHF 1-2 | Easy |
| Consolidate to single environment | CHF 12-15 | High |
| Optimize backup retention | CHF 0.5-1 | Easy |

**Total potential savings: CHF 10-14/month** (enough to meet target!)

### Detailed Non-Destructive Strategies

#### 1. Switch to GitHub Container Registry (Recommended)

**Savings: CHF 4-6/month** (eliminates ACR costs entirely)

GitHub Container Registry (GHCR) is free and works seamlessly with Container Apps:

**Implementation:**
1. Update build workflows to push to `ghcr.io` instead of ACR
2. Update Container Apps to pull from GHCR with GitHub token
3. Once verified, delete ACR resources

**Steps:**
```yaml
# In .github/workflows/10-build-push-images.yml
- name: Login to GitHub Container Registry
  uses: docker/login-action@v3
  with:
    registry: ghcr.io
    username: ${{ github.actor }}
    password: ${{ secrets.GITHUB_TOKEN }}

- name: Build and push
  uses: docker/build-push-action@v5
  with:
    push: true
    tags: ghcr.io/${{ github.repository }}/frontend:${{ github.ref_name }}
```

Then update Terraform to use GHCR:
```hcl
# In bfs-cloud/envs/production/main.tf
registries = [
  {
    server               = "ghcr.io"
    username             = "ly-schneider"  # GitHub username
    password_secret_name = "ghcr-token"    # Add GitHub PAT as secret
  }
]
```

**Pros:**
- No ACR costs
- Free unlimited public images
- Private images free for reasonable usage
- Built-in image scanning

**Cons:**
- Requires GitHub token management
- Not Azure-native

#### 2. Share Container Registry Between Environments

**Savings: CHF 4-6/month** (one ACR instead of two)

Create a single shared ACR for both staging and production:

**Implementation:**
1. Create `bfs-shared-acr` in a separate resource group
2. Configure both environments to use the shared ACR
3. Delete environment-specific ACRs
4. Use tags to separate staging/production images

**Configuration:**
```hcl
# Create shared ACR once (manually or in separate Terraform)
# Then in both envs/staging/main.tf and envs/production/main.tf:

module "bfs_infrastructure" {
  # ...
  config = {
    enable_acr = false  # Don't create environment-specific ACR
    acr_login_server = "bfssharedacr.azurecr.io"
    acr_resource_id = "/subscriptions/.../bfssharedacr"
    # ...
  }
}
```

**Pros:**
- 50% ACR cost reduction
- Simplified image management
- Easier to share base images

**Cons:**
- Less environment isolation
- Requires coordination between environments

#### 3. Reduce Cosmos DB Diagnostic Logging

**Savings: CHF 1-2/month**

Cosmos DB has extensive diagnostic logging that generates Log Analytics costs. The configuration currently enables 4 log categories that may not be needed.

**Current configuration** (in `bfs-cloud/modules/cosmos_mongo/main.tf`):
- DataPlaneRequests
- MongoRequests
- QueryRuntimeStatistics
- PartitionKeyStatistics

**Optimization:**
Keep only essential logs for production monitoring:

```hcl
# Only enable essential logging
enabled_log {
  category = "MongoRequests"  # Keep this for error tracking
}

# Comment out or remove:
# - DataPlaneRequests (verbose, use only for debugging)
# - QueryRuntimeStatistics (use only for performance tuning)
# - PartitionKeyStatistics (use only for optimization work)
```

**Pros:**
- Immediate cost reduction
- Cleaner logs
- Faster query performance in Log Analytics

**Cons:**
- Less diagnostic data if issues occur
- May need to re-enable for troubleshooting

#### 4. Optimize Backup Configuration

**Savings: CHF 0.5-1/month**

Cosmos DB backup is set to Geo-redundant with frequent intervals:

**Current settings:**
- Type: Periodic
- Interval: 240 minutes (4 hours)
- Retention: 8 hours
- Redundancy: Geo

**Optimization for non-critical environments:**
```hcl
backup {
  type                = "Periodic"
  interval_in_minutes = 1440  # Daily instead of 4-hourly
  retention_in_hours  = 8     # Keep
  storage_redundancy  = "Local"  # Local instead of Geo
}
```

For staging environment, consider continuous backup only when needed:
```hcl
backup {
  type = "Continuous"  # Only pay when restoring
}
```

**Pros:**
- Lower storage costs
- Still have disaster recovery
- Continuous backup available if needed

**Cons:**
- Less frequent backups
- No geo-redundancy (but cheaper to replicate manually if needed)

#### 5. Consolidate to Single Environment

**Savings: CHF 12-15/month** (most significant!)

If you can operate with just production and use feature flags/branches for testing:

**What to delete:**
- Entire staging environment (Container Apps Environment, VNet, Cosmos DB, etc.)

**Alternative testing approach:**
- Use local development with Docker Compose
- Use preview environments (temporary, on-demand)
- Use feature flags for gradual rollouts

**Pros:**
- Cuts infrastructure costs nearly in half
- Simpler to manage
- Forces better development practices

**Cons:**
- No permanent staging environment
- Need robust local development setup
- Higher risk deployments

## Solution: Destroy and Recreate Infrastructure (Maximum Savings)

If you need the absolute lowest costs and can tolerate downtime, destroying infrastructure when not in use is the most effective approach.

### Cost Breakdown

| Scenario | Monthly Cost |
|----------|-------------|
| Infrastructure always running | ~CHF 25 |
| Apps scaled to 0 (current) | ~CHF 16-27 |
| With optimizations (no deletion) | ~CHF 11-17 |
| Infrastructure destroyed | ~CHF 0-2* |

*Small costs from Azure AD, networking remnants, or storage may persist

### Manual Destroy/Deploy Process

When the app won't be used for several days:

1. **Destroy infrastructure** using Terraform Cloud:
   - Go to https://app.terraform.io/app/leys-services/workspaces/bfs-production
   - Click Settings → Destruction and Deletion
   - Click "Queue destroy plan"
   - Review and confirm

2. **Redeploy when needed**:
   - Push code to production branch OR
   - Manually trigger deployment workflow in GitHub Actions

**Pros:**
- Near-zero costs when destroyed (saves ~CHF 20-25/month)
- Full control over when infrastructure exists

**Cons:**
- Manual process (5-10 minutes to destroy, 10-15 minutes to recreate)
- Apps unavailable during destroyed period
- Need to recreate on first use

## Implementation Guide

### Destroying Infrastructure

**Using Terraform Cloud UI:**
```
1. Navigate to workspace: https://app.terraform.io/app/leys-services/workspaces/bfs-production
2. Settings → Destruction and Deletion
3. Queue destroy plan
4. Review resources to be destroyed
5. Confirm destruction
```

**Using GitHub Actions:**
```
1. Go to Actions tab in GitHub
2. Select "Destroy Infrastructure" workflow
3. Click "Run workflow"
4. Select environment (staging/production)
5. Type "DESTROY" to confirm
6. Click "Run workflow"
```

**Using Terraform CLI:**
```bash
cd bfs-cloud/envs/production
terraform destroy
```

### Recreating Infrastructure

Infrastructure will be automatically recreated when:
- Code is pushed to the production branch
- A deployment workflow is manually triggered

**Manual deployment:**
```
1. Go to GitHub Actions
2. Select deployment workflow
3. Click "Run workflow"
4. Select environment
5. Click "Run workflow"
```

## Recommended Strategy

### For Continuous Availability (No Downtime)

**Target: CHF 11-15/month** (down from CHF 25)

Apply these optimizations in order of impact:

1. **Switch to GitHub Container Registry** (CHF 4-6 saved)
   - Easiest and most impactful
   - No infrastructure changes needed
   - Free alternative to ACR

2. **Reduce Cosmos DB diagnostic logging** (CHF 1-2 saved)
   - Quick Terraform config change
   - Remove verbose logs you don't need

3. **Optimize backup settings** (CHF 0.5-1 saved)
   - Change to daily backups
   - Use local redundancy for staging

4. **Consider consolidating environments** (CHF 12-15 saved)
   - If you can test locally
   - Biggest single saving without deletion

**Result:** Meet target without infrastructure destruction!

### For Intermittent Usage (Can Tolerate Downtime)

**Target: CHF 0-5/month** (near zero)

For production environment with intermittent usage:

1. **Keep infrastructure running during active periods**
   - Events, regular usage, testing
   
2. **Destroy infrastructure during quiet periods**
   - Evenings/nights if not used
   - Weekends if not needed
   - Holiday periods
   - Between events

3. **Quick restore capability**
   - Can redeploy in 10-15 minutes
   - All configuration in code

**Result:** Maximum savings with acceptable downtime.

3. **Quick restore capability**
   - Can redeploy in 10-15 minutes
   - All configuration in code
   - No manual setup required

## Cost Monitoring

Set up budget alerts to track actual costs:

```hcl
budget_amount = 10  # Alert when approaching CHF 10/month
```

Configure in `bfs-cloud/envs/production/main.tf`:
- Budget alerts trigger at 80%, 100%, and 120% of threshold
- Email notifications to configured addresses

## Alternative Architecture (Future)

For truly on-demand, consider:

1. **Azure Container Instances**
   - Pay only when running
   - No environment base cost
   - Less features than Container Apps

2. **Azure Functions**
   - Consumption plan (truly serverless)
   - Pay per execution
   - May require app restructuring

3. **Azure Static Web Apps** + **Azure Functions**
   - Frontend: Static Web Apps (free tier available)
   - Backend: Functions (consumption plan)
   - Near-zero idle costs
