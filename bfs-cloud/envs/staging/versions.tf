terraform {
  required_version = ">= 1.6.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 4.1.0"
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
