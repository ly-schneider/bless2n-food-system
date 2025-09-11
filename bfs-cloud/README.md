# Bless2n Food System (BFS) — Azure Terraform

## What this deploys
- Per‐environment isolation: RG, VNet, delegated Subnet, Container Apps Environment, Container Apps, Cosmos DB (Mongo API).
- Observability: Log Analytics (both envs), workspace-based Application Insights (prod), Diagnostic Settings at env/app level, sample 5xx metric alerts.

## Usage
```bash
cd terraform/envs/prod
terraform init
terraform apply -auto-approve
# Repeat in ./staging
```