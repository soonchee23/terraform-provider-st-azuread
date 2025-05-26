Terraform Custom Provider for AzureAD
======================================

This Terraform custom provider is designed for own use case scenario.

Supported Versions
------------------

| Terraform version | minimum provider version |maxmimum provider version
| ---- | ---- | ----|
| >= 1.12.x	| 0.1.1	| latest |

Requirements
------------

-	[Terraform](https://www.terraform.io/downloads.html) 1.12.x
-	[Go](https://golang.org/doc/install) 1.23 (to build the provider plugin)

Local Installation
------------------

1. Run make file `make install-local-custom-provider` to install the provider under ~/.terraform.d/plugins.

2. The provider source should be change to the path that configured in the *Makefile*:

    ```
    terraform {
      required_providers {
        st-azuread = {
          source = "example.local/myklst/st-azuread"
        }
      }
    }

    provider "st-azuread" {}
    ```

Why Custom Provider
-------------------

This custom provider exists due to some of the resources and data sources in the
official AzureAD Terraform provider may not fulfill the requirements of some
scenario. The reason behind every resources and data sources are stated as below:

### Resources

- **st-azuread_named_location**

  - The official AzureAD Terraform provider's [*azuread_named_location*](https://registry.terraform.io/providers/hashicorp/azuread/3.3.0/docs/resources/named_location) resource may encounter API rate limiting issues when provisioning multiple resources concurrently.
  To enhance reliability and scalability, a backoff retry mechanism has been integrated, ensuring successful creation
  of multiple resources simultaneously.

### Data Sources

- **st-azuread_auth_strength_policy**

  - The official AzureAD Terraform provider does not currently support retrieving authentication strength policies in
  Microsoft Entra ID by ID or name. This data source able to support retrieving authentication strength policies in
  Microsoft Entra ID by ID or name.

- **st-azuread_groups**

  - The official AzureAD Terraform provider's [*azuread_groups*](https://registry.terraform.io/providers/hashicorp/azuread/3.3.0/docs/data-sources/groups)  data source separates display_names and object_ids into separate lists, which might lead to potential errors.
  This data source uses a list of objects for groups, ensuring that each group's display_name is directly associated with
  its corresponding id, thereby reducing the risk of mismatches and improving data integrity.

References
----------

- Website: https://www.terraform.io
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework
- AzureAD official Terraform provider: https://github.com/hashicorp/terraform-provider-azuread
