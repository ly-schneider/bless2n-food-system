module "bfs_infrastructure" {
  source = "../../common"

  environment   = "staging"
  location      = var.location
  tags          = var.tags
  images        = var.images
  alert_emails  = var.alert_emails

  config = {
    rg_name                = "bfs-staging-rg"
    vnet_name              = "bfs-staging-vnet"
    subnet_name            = "container-apps-subnet"
    vnet_cidr              = "10.1.0.0/16"
    subnet_cidr            = "10.1.0.0/21"
    env_name               = "bfs-staging-env"
    law_name               = "bfs-logs-workspace"
    appi_name              = "bfs-staging-insights"
    enable_app_insights    = false
    retention_days         = 14
    cosmos_name            = "bfs-staging-cosmos"
    database_throughput    = 400
    enable_alerts          = false
    requests_5xx_threshold = 10
    
    apps = {
      frontend = {
        port         = 80
        image        = var.images.frontend_staging_01
        cpu          = 0.25
        memory       = "0.5Gi"
        min_replicas = 0
        max_replicas = 20
        http_scale_rule = {
          name                = "frontend-http-scale"
          concurrent_requests = 20
        }
        cpu_scale_rule = {
          name           = "frontend-cpu-scale"
          cpu_percentage = 75
        }
      }
      backend = {
        port         = 8080
        image        = var.images.backend_staging_01
        cpu          = 0.5
        memory       = "1.0Gi"
        min_replicas = 0
        max_replicas = 20
        http_scale_rule = {
          name                = "backend-http-scale"
          concurrent_requests = 40
        }
        cpu_scale_rule = {
          name           = "backend-cpu-scale"
          cpu_percentage = 80
        }
      }
    }
  }
}