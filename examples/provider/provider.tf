terraform {
  required_providers {
    st-azuread = {
      source = "myklst/st-azuread"
    }
  }
}

provider "azuread" {
  client_id     = "..."
  client_secret = "..."
  tenant_id     = "..."
}
