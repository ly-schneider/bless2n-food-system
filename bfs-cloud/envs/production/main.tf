module "config" {
  source = "../../defaults"

  env             = "production"
  image_tag       = var.image_tag
  frontend_digest = var.frontend_digest
  backend_digest  = var.backend_digest
  revision_suffix = var.revision_suffix
  ghcr_token      = var.ghcr_token

  frontend_cpu    = 0.5
  frontend_memory = "1Gi"

  backend_cpu    = 0.5
  backend_memory = "1Gi"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment = "production"
  location    = module.config.location
  tags        = module.config.tags
  config      = module.config.config
}
