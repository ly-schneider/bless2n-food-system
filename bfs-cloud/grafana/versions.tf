terraform {
  required_version = ">= 1.13.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "4.68.0"
    }
    azuread = {
      source  = "hashicorp/azuread"
      version = "3.8.0"
    }
    grafana = {
      source  = "grafana/grafana"
      version = "~> 3.18"
    }
  }

  cloud {
    organization = "leys-services"

    workspaces {
      name = "bfs-grafana"
    }
  }
}

provider "azurerm" {
  features {}
}

provider "azuread" {}

provider "grafana" {
  url  = var.grafana_stack_url
  auth = var.grafana_cloud_token
}
