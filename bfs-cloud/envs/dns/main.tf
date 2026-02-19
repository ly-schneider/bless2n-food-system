resource "azurerm_resource_group" "dns" {
  name     = "bfs-dns-rg"
  location = "northeurope"
}

resource "azurerm_dns_zone" "food" {
  name                = "food.blessthun.ch"
  resource_group_name = azurerm_resource_group.dns.name
}
