import {
  id = "/subscriptions/5c3f67c0-445b-4317-a959-e9d5c663dc16/resourceGroups/bfs-staging-rg/providers/Microsoft.App/containerApps/backend-staging"
  to = module.bfs_infrastructure.module.apps["backend-staging"].azurerm_container_app.this
}

import {
  id = "/subscriptions/5c3f67c0-445b-4317-a959-e9d5c663dc16/resourceGroups/bfs-staging-rg/providers/Microsoft.App/containerApps/frontend-staging"
  to = module.bfs_infrastructure.module.apps["frontend-staging"].azurerm_container_app.this
}
