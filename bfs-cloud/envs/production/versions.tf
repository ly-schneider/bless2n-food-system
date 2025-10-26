terraform {
  required_version = ">= 1.6.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = ">= 4.23.0, < 5.0.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = ">= 2.47.0"
    }
  }

  cloud {
    organization = "leys-services"

    workspaces {
      name = "bfs-production"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}
