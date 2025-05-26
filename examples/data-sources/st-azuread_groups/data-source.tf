data "st-azuread_groups" "example" {
  display_names = ["group-a", "group-b"]
}

data "st-azuread_groups" "sales" {
  display_name_prefix = "sales-"
}

data "st-azuread_groups" "all" {
  return_all = true
}

data "st-azuread_groups" "mail_enabled" {
  mail_enabled = true
  return_all   = true
}

data "st-azuread_groups" "security_only" {
  mail_enabled     = false
  return_all       = true
  security_enabled = true
}
