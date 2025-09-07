output "account_id"   { value = azurerm_cosmosdb_account.this.id }
output "account_name" { value = azurerm_cosmosdb_account.this.name }
output "db_name"      { value = try(azurerm_cosmosdb_mongo_database.db[0].name, null) }