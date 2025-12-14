terraform {
  required_version = ">= 1.13.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.56.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "3.7.0"
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
