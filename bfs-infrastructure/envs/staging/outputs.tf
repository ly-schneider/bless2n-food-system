output "resource_group" { value = module.rg.name }
output "aca_env_name"   { value = module.aca_env.name }
output "log_analytics"  { value = module.obs.log_analytics_name }
output "cosmos_account" { value = module.cosmos.account_name }