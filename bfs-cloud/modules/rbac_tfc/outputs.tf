output "principal_object_id" {
  description = "Resolved object ID of the principal that received role assignments"
  value       = local.principal_object_id
}

output "assigned_roles" {
  description = "Counts of assignments by category"
  value = {
    baseline_contributor = length(azurerm_role_assignment.baseline_contributor_rg)
    baseline_uaa         = length(azurerm_role_assignment.baseline_uaa_rg)
    network              = length(azurerm_role_assignment.network_contributor)
    private_dns          = length(azurerm_role_assignment.private_dns_zone_contributor)
    managed_identity     = length(azurerm_role_assignment.managed_identity_contributor)
    uaa                  = length(azurerm_role_assignment.uaa)
    cosmos_contributor   = length(azurerm_role_assignment.cosmos_account_contributor)
    cosmos_keys_reader   = length(azurerm_role_assignment.cosmos_keys_reader)
  }
}

