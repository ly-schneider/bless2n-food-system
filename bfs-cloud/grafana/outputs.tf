output "grafana_folder_url" {
  description = "URL to the Bless2n Food System folder in Grafana"
  value       = "${var.grafana_stack_url}/dashboards/f/${grafana_folder.bfs.uid}"
}

output "dashboard_url" {
  description = "URL to the main overview dashboard"
  value       = "${var.grafana_stack_url}${grafana_dashboard.overview.url}"
}

output "service_principal_client_id" {
  description = "Client ID of sp-grafana-cloud-reader (used by Azure Monitor data source)"
  value       = azuread_application.grafana.client_id
}

output "service_principal_object_id" {
  description = "Object ID of sp-grafana-cloud-reader (for auditing role assignments)"
  value       = azuread_service_principal.grafana.object_id
}
