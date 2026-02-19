module "config" {
  source = "../../defaults"

  env             = "staging"
  image_tag       = var.image_tag
  frontend_digest = var.frontend_digest
  backend_digest  = var.backend_digest
  revision_suffix = var.revision_suffix
  ghcr_token      = var.ghcr_token
  docs_digest     = var.docs_digest
  domain_prefix   = "staging"
}

module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment = "staging"
  location    = module.config.location
  tags        = module.config.tags
  config      = module.config.config
  dns         = module.config.dns
}
