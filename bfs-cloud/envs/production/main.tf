module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment = "production"
  location    = module.config.location
  tags        = module.config.tags
  config      = module.config.config
}
