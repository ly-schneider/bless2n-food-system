module "config" {
  source = "../../defaults"

  env             = "production"
  image_tag       = var.image_tag
  frontend_digest = var.frontend_digest
  backend_digest  = var.backend_digest
  revision_suffix = var.revision_suffix
  ghcr_token      = var.ghcr_token
  app_secrets     = var.app_secrets

  # Production frontend gets more resources
  frontend_cpu    = 0.5
  frontend_memory = "1Gi"

  # Production API keys
  frontend_stripe_publishable_key = "pk_live_51PBsTW2LfBQkI29zWnUaY1HAsP34VeFoYsInuhNNpyHjrZiHhrDZrMGvCIzZuZOXJEzxpOXhQyW8wfi6qiwelHkh00TTmNdU4k"
  frontend_google_client_id       = "728225904671-bsdo738sald74qkbg38etqjqj5jjfh66.apps.googleusercontent.com"
  backend_google_client_id        = "728225904671-bsdo738sald74qkbg38etqjqj5jjfh66.apps.googleusercontent.com"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment = "production"
  location    = module.config.location
  tags        = module.config.tags
  config      = module.config.config
}
