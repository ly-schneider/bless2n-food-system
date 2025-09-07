variable "location" {
  type    = string
  default = "westeurope"
}
variable "tags" {
  type    = map(string)
  default = { project = "bfs" }
}
variable "alert_emails" {
  type    = map(string)
  default = {}
}

variable "images" {
  type = map(string)
  default = {
    frontend_prod_01 = "nginx:1.25"
    frontend_prod_02 = "nginx:1.25"
    backend_prod_01  = "ghcr.io/example/backend:latest"
    backend_prod_02  = "ghcr.io/example/backend:latest"
  }
}