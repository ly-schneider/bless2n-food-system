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

## Solution: Destroy and Recreate Infrastructure

To achieve costs under CHF 10/month, the infrastructure must be **completely destroyed** when not in use.

### Cost Breakdown

| Scenario | Monthly Cost |
|----------|-------------|
| Infrastructure always running | ~CHF 25 |
| Apps scaled to 0 (current) | ~CHF 16-27 |
| Infrastructure destroyed | ~CHF 0-2* |
| Infrastructure destroyed + shared ACR | ~CHF 0-1* |

*Small costs from Azure AD, networking remnants, or storage may persist

### Option 1: Manual Destroy/Deploy (Recommended)

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

### Option 2: Share Container Registry

Both staging and production can share a single ACR to reduce costs:

1. Create one ACR in a shared resource group
2. Configure both environments to use the shared ACR
3. Delete environment-specific ACRs

**Savings:** ~CHF 4-6/month (one ACR instead of two)

### Option 3: Use GitHub Container Registry

Replace Azure Container Registry with GitHub Container Registry (free):

1. Build and push images to ghcr.io instead of ACR
2. Configure Container Apps to pull from GHCR
3. Destroy Azure Container Registry

**Savings:** ~CHF 4-6/month

### Option 4: Optimize Log Analytics

Reduce Log Analytics retention and data collection:

1. Set retention to minimum (30 days)
2. Disable Application Insights if not needed
3. Reduce diagnostic settings to essential logs only

**Savings:** ~CHF 1-3/month

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
