output "vnet_id" { value = azurerm_virtual_network.this.id }
output "subnet_id" { value = azurerm_subnet.aca.id }
output "subnet_address_prefix" { value = var.subnet_cidr }
output "workload_profile_subnet_id" { value = azurerm_subnet.aca_workload_profiles.id }
output "workload_profile_subnet_prefix" { value = var.workload_profiles_subnet_cidr }
output "private_endpoints_subnet_id" { value = azurerm_subnet.private_endpoints.id }
output "private_endpoints_subnet_prefix" { value = var.private_endpoints_subnet_cidr }
