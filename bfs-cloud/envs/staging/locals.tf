module "config" {
  source = "../../defaults"

  env             = "staging"
  image_tag       = var.image_tag
  frontend_digest = var.frontend_digest
  backend_digest  = var.backend_digest
  revision_suffix = var.revision_suffix
  ghcr_token      = var.ghcr_token
  app_secrets     = var.app_secrets

  frontend_stripe_publishable_key = "pk_test_51RQl9ZRqpUl5qfpdqiu8nhU6h5N1YyEXWoOxqUsPb8UouZqPMubOZtESdFa4KHTWM71GhAbbddlS3a6aTFu1vIDe00p1DqQTG9"
  frontend_google_client_id       = "728225904671-h9cp0badsuvamscrn6k2lnkksiinld99.apps.googleusercontent.com"
  backend_google_client_id        = "728225904671-h9cp0badsuvamscrn6k2lnkksiinld99.apps.googleusercontent.com"
}
