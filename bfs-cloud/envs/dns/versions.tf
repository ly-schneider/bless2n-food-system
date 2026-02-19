terraform {
  required_version = ">= 1.13.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.60.0"
    }
  }

  cloud {
    organization = "leys-services"

    workspaces {
      name = "bfs-dns"
    }
  }
}

provider "azurerm" {
  features {}
}
