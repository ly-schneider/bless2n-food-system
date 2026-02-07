# This file removes Key Vault resources from Terraform state without destroying them.
# After running `terraform apply`, the Key Vault and its secrets will no longer be
# managed by Terraform but will continue to exist in Azure.
#
# Once applied successfully, you can delete this file and the original resource
# definitions from main.tf.

removed {
  from = azurerm_key_vault.basic

  lifecycle {
    destroy = false
  }
}

removed {
  from = azurerm_key_vault_access_policy.terraform_admin

  lifecycle {
    destroy = false
  }
}

removed {
  from = azurerm_key_vault_access_policy.container_apps_identity

  lifecycle {
    destroy = false
  }
}

removed {
  from = azurerm_monitor_diagnostic_setting.basic_security_logs

  lifecycle {
    destroy = false
  }
}
