output "grafana_folder_url" {
  description = "URL to the Bless2n Food System folder in Grafana"
  value       = "${var.grafana_stack_url}/dashboards/f/${grafana_folder.bfs.uid}"
}

output "dashboard_urls" {
  description = "URLs to the per-env overview dashboards"
  value       = { for env, d in grafana_dashboard.overview : env => "${var.grafana_stack_url}${d.url}" }
}
