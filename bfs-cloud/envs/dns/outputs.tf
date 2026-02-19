output "zone_name" {
  value = azurerm_dns_zone.food.name
}

output "resource_group_name" {
  value = azurerm_resource_group.dns.name
}

output "name_servers" {
  value = azurerm_dns_zone.food.name_servers
}
