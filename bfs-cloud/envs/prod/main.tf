module "bfs_infrastructure" {
  source = "../../common"

  environment   = "prod"
  location      = var.location
  tags          = var.tags
  images        = var.images
  alert_emails  = var.alert_emails

  config = {
    rg_name                      = "bfs-prod-rg"
    vnet_name                    = "bfs-prod-vnet"
    subnet_name                  = "container-apps-subnet"
    vnet_cidr                    = "10.0.0.0/16"
    subnet_cidr                  = "10.0.0.0/21"
    env_name                     = "bfs-prod-env"
    law_name                     = "bfs-logs-workspace"
    appi_name                    = "bfs-prod-insights"
    enable_app_insights          = true
    retention_days               = 30
    cosmos_name                  = "bfs-prod-cosmos"
    database_throughput          = 400
    enable_alerts                = true
    requests_5xx_threshold       = 10
    enable_security_features     = true
    key_vault_name               = "bfs-prod-kv"
    
    apps = {
      frontend = {
        port         = 80
        image        = var.images.frontend_prod_01
        cpu          = 0.5
        memory       = "1.0Gi"
        min_replicas = 0
        max_replicas = 20
        http_scale_rule = {
          name                = "frontend-http-scale"
          concurrent_requests = 30
        }
        cpu_scale_rule = {
          name           = "frontend-cpu-scale"
          cpu_percentage = 70
        }
      }
      backend = {
        port         = 8080
        image        = var.images.backend_prod_01
        cpu          = 1.0
        memory       = "2.0Gi"
        min_replicas = 0
        max_replicas = 20
        http_scale_rule = {
          name                = "backend-http-scale"
          concurrent_requests = 50
        }
        cpu_scale_rule = {
          name           = "backend-cpu-scale"
          cpu_percentage = 80
        }
        memory_scale_rule = {
          name              = "backend-memory-scale"
          memory_percentage = 85
        }
      }
    }
  }
}