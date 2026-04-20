output "grafana_folder_url" {
  description = "URL to the Bless2n Food System folder in Grafana"
  value       = "${var.grafana_stack_url}/dashboards/f/${grafana_folder.bfs.uid}"
}

output "dashboard_url" {
  description = "URL to the main overview dashboard"
  value       = "${var.grafana_stack_url}${grafana_dashboard.overview.url}"
}
