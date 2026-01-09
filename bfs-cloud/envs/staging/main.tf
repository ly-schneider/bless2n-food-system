module "bfs_infrastructure" {
  source = "../../modules/stack"

  environment = "staging"
  location    = module.config.location
  tags        = module.config.tags
  config      = module.config.config
}
