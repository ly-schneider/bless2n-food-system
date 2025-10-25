output "resource_group_name" {
  description = "Name of the created resource group"
  value       = module.bfs_infrastructure.resource_group_name
}

output "vnet_id" {
  description = "ID of the virtual network"
  value       = module.bfs_infrastructure.vnet_id
}

output "subnet_id" {
  description = "ID of the container apps subnet"
  value       = module.bfs_infrastructure.subnet_id
}

output "container_app_environment_id" {
  description = "ID of the container app environment"
  value       = module.bfs_infrastructure.container_app_environment_id
}

output "log_analytics_workspace_id" {
  description = "ID of the Log Analytics workspace"
  value       = module.bfs_infrastructure.log_analytics_workspace_id
}

output "app_insights_connection_string" {
  description = "Application Insights connection string"
  value       = module.bfs_infrastructure.app_insights_connection_string
  sensitive   = true
}

output "cosmos_connection_string" {
  description = "Cosmos DB connection string"
  value       = module.bfs_infrastructure.cosmos_connection_string
  sensitive   = true
}

output "app_urls" {
  description = "URLs of the deployed applications"
  value       = module.bfs_infrastructure.app_urls
}