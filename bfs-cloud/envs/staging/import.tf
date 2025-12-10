// Import existing diagnostic setting for backend-staging so Terraform manages it.
import {
  to = module.bfs_infrastructure.module.apps_backend["backend-staging"].module.diag.azurerm_monitor_diagnostic_setting.this
  id = "/subscriptions/a6e97570-b80b-4a4d-a27c-081fe6c7e7f3/resourceGroups/bfs-staging-rg/providers/Microsoft.App/containerApps/backend-staging|backend-staging-diag"
}
