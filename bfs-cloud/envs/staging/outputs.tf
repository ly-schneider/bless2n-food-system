output "resource_group_name" {
  description = "Name of the env resource group"
  value       = module.bfs_infrastructure.resource_group_name
}

output "log_analytics_workspace_id" {
  description = "Azure resource ID of the Log Analytics workspace"
  value       = module.bfs_infrastructure.log_analytics_workspace_id
}
