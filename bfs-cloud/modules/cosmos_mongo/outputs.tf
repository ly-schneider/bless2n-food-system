output "account_id" {
  description = "Cosmos DB account ID"
  value       = azurerm_cosmosdb_account.this.id
}

output "account_name" {
  description = "Cosmos DB account name"
  value       = azurerm_cosmosdb_account.this.name
}

output "db_name" {
  description = "MongoDB database name"
  value       = try(azurerm_cosmosdb_mongo_database.db[0].name, null)
}

output "connection_string" {
  description = "Primary MongoDB connection string"
  value       = azurerm_cosmosdb_account.this.primary_mongodb_connection_string
  sensitive   = true
}
