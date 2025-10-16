terraform {
  required_version = ">= 1.6.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      # Pin to the minor version that originally created state to avoid
      # provider/state decode issues observed with later 4.x (e.g., value_wo, enabled_metric).
      # Upgrade incrementally after a successful plan.
      version = "= 4.9.0"
    }
  }

  cloud {
    organization = "leys-services"

    workspaces {
      name = "bfs-staging"
    }
  }
}

provider "azurerm" {
  features {}
}
