# Bless2n Food System (BFS) — Azure Terraform

## Architecture Overview

This Terraform infrastructure uses a **unified state management** approach with environment-specific configurations. All infrastructure logic is centralized in the `common/` module, while environments differ only through variable configurations.

## What this deploys

### Infrastructure Components
- **Per-environment isolation**: Resource Groups, VNets, delegated Subnets, Container Apps Environment
- **Container Apps**: Auto-scaling web applications with intelligent scaling rules
- **Database**: Cosmos DB with MongoDB API
- **Observability**: Log Analytics (both envs), Application Insights (prod), Diagnostic Settings
- **Monitoring**: Metric-based alerts and monitoring (configurable per environment)

### Auto-Scaling Configuration
- **Scale-to-Zero**: Applications scale down to 0 replicas when idle (cost-efficient)
- **Burst Scaling**: Can scale up to 20 replicas during traffic spikes
- **Smart Triggers**: HTTP request-based, CPU percentage, and memory percentage scaling
- **Environment-Specific**: Different scaling thresholds for prod vs staging

## Project Structure

```
bfs-cloud/
├── common/                    # Unified infrastructure module
│   ├── main.tf               # Core infrastructure logic
│   ├── outputs.tf            # Shared outputs
│   └── env-configs/          # Optional: Environment-specific tfvars
│       ├── prod.tfvars
│       └── staging.tfvars
├── envs/
│   ├── prod/                 # Production environment
│   │   ├── main.tf          # Calls common module with prod config
│   │   ├── variables.tf     # Prod-specific variables
│   │   ├── outputs.tf       # Prod outputs
│   │   └── backend.tf       # Terraform backend config
│   └── staging/             # Staging environment
│       ├── main.tf          # Calls common module with staging config
│       ├── variables.tf     # Staging-specific variables
│       ├── outputs.tf       # Staging outputs
│       └── backend.tf       # Terraform backend config
└── modules/                 # Reusable Terraform modules
    ├── containerapp/        # Enhanced with auto-scaling rules
    ├── network/
    ├── observability/
    └── ...
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
- **Environment-Appropriate**: Different thresholds for prod vs staging

## Usage

### Standard Deployment
```bash
# Deploy production
cd envs/prod
terraform init
terraform apply

# Deploy staging  
cd envs/staging
terraform init
terraform apply
```

### Using Environment Configs (Alternative)
```bash
# Deploy using tfvars files
cd common
terraform init
terraform apply -var-file="env-configs/prod.tfvars"
terraform apply -var-file="env-configs/staging.tfvars"
```

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