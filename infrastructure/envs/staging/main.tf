locals {
  rg_name     = "bfs-staging-rg"
  vnet_name   = "bfs-staging-vnet"
  subnet_name = "container-apps-subnet"
  vnet_cidr   = "10.1.0.0/16"
  subnet_cidr = "10.1.0.0/21"

  env_name    = "bfs-staging-env"
  law_name    = "bfs-logs-workspace"

  apps = {
    frontend-staging-01 = { port = 80,   image = var.images.frontend_staging_01 }
    backend-staging-01  = { port = 8080, image = var.images.backend_staging_01 }
  }
}

module "rg" {
  source   = "../../modules/rg"
  name     = local.rg_name
  location = var.location
  tags     = var.tags
}

module "net" {
  source              = "../../modules/network"
  resource_group_name = module.rg.name
  location            = var.location
  vnet_name           = local.vnet_name
  vnet_cidr           = local.vnet_cidr
  subnet_name         = local.subnet_name
  subnet_cidr         = local.subnet_cidr
  tags                = var.tags
}

module "obs" {
  source               = "../../modules/observability"
  resource_group_name  = module.rg.name
  location             = var.location
  law_name             = local.law_name
  appi_name            = "bfs-staging-insights"
  enable_app_insights  = false
  retention_days       = 14
  tags                 = var.tags
}

module "aca_env" {
  source                      = "../../modules/containerapps_env"
  name                        = local.env_name
  location                    = var.location
  resource_group_name         = module.rg.name
  subnet_id                   = module.net.subnet_id
  logs_destination            = "azure-monitor"
  log_analytics_workspace_id  = module.obs.log_analytics_id
  tags                        = var.tags
}

module "env_diag" {
  source                      = "../../modules/diagnostic_setting"
  name                        = "${local.env_name}-diag"
  target_resource_id          = module.aca_env.id
  log_analytics_workspace_id  = module.obs.log_analytics_id
  categories                  = []
  category_groups             = ["allLogs"]
  enable_metrics              = true
}

module "cosmos" {
  source              = "../../modules/cosmos_mongo"
  name                = "bfs-staging-cosmos"
  location            = var.location
  resource_group_name = module.rg.name
  create_database     = true
  database_name       = "appdb"
  database_throughput = 400
  tags                = var.tags
}

module "apps" {
  for_each = local.apps
  source   = "../../modules/containerapp"

  name                        = each.key
  resource_group_name         = module.rg.name
  environment_id              = module.aca_env.id
  image                       = each.value.image
  target_port                 = each.value.port
  cpu                         = 0.5
  memory                      = "1.0Gi"
  min_replicas                = 1
  max_replicas                = 2
  enable_system_identity      = true
  log_analytics_workspace_id  = module.obs.log_analytics_id
  environment_variables       = {}
  tags                        = merge(var.tags, { app = each.key })
}