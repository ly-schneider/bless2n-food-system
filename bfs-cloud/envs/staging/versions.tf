terraform {
  required_version = ">= 1.13.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.52.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "3.6.0"
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

provider "azuread" {}
