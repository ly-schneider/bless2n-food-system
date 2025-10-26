data "azurerm_client_config" "current" {}

# Optionally resolve principal by client_id (appId) via AzureAD; otherwise use current principal
data "azuread_service_principal" "target" {
  count     = var.principal_object_id == null && var.principal_client_id != null && var.principal_client_id != "" ? 1 : 0
  client_id = var.principal_client_id
}

locals {
  principal_object_id = coalesce(
    var.principal_object_id,
    try(data.azuread_service_principal.target[0].object_id, null),
    data.azurerm_client_config.current.object_id
  )
}

# Baseline (optional): RG-scoped Contributor and/or UAA
resource "azurerm_role_assignment" "baseline_contributor_rg" {
  count               = var.baseline_enable_contributor_on_rg ? 1 : 0
  scope               = var.target_rg_id
  role_definition_name = "Contributor"
  principal_id        = local.principal_object_id
}

resource "azurerm_role_assignment" "baseline_uaa_rg" {
  count               = var.baseline_enable_uaa_on_rg ? 1 : 0
  scope               = var.target_rg_id
  role_definition_name = "User Access Administrator"
  principal_id        = local.principal_object_id
}

# Network: Network Contributor on provided scopes
resource "azurerm_role_assignment" "network_contributor" {
  for_each            = toset(var.network_scopes)
  scope               = each.value
  role_definition_name = "Network Contributor"
  principal_id        = local.principal_object_id
}

# Private DNS: Private DNS Zone Contributor on provided scopes
resource "azurerm_role_assignment" "private_dns_zone_contributor" {
  for_each            = toset(var.private_dns_zone_scopes)
  scope               = each.value
  role_definition_name = "Private DNS Zone Contributor"
  principal_id        = local.principal_object_id
}

# Managed Identity: allow creating user-assigned identities when needed
resource "azurerm_role_assignment" "managed_identity_contributor" {
  for_each            = toset(var.managed_identity_scopes)
  scope               = each.value
  role_definition_name = "Managed Identity Contributor"
  principal_id        = local.principal_object_id
}

# RBAC self-grant: allow Terraform to create role assignments where required
resource "azurerm_role_assignment" "uaa" {
  for_each            = toset(var.uaa_scopes)
  scope               = each.value
  role_definition_name = "User Access Administrator"
  principal_id        = local.principal_object_id
}

# Cosmos: management contributor for create/update (least-privilege vs broad Contributor)
resource "azurerm_role_assignment" "cosmos_account_contributor" {
  for_each            = var.grant_cosmos_account_contributor ? toset(var.cosmos_account_scopes) : []
  scope               = each.value
  role_definition_name = var.cosmos_management_role_definition_name
  principal_id        = local.principal_object_id
}

# Cosmos: keys read without wider writes
resource "azurerm_role_assignment" "cosmos_keys_reader" {
  for_each            = var.grant_cosmos_keys_reader ? toset(var.cosmos_account_scopes) : []
  scope               = each.value
  role_definition_name = "Cosmos DB Account Reader Role"
  principal_id        = local.principal_object_id
}
