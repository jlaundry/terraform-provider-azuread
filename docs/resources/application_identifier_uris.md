---
subcategory: "Applications"
---

# Resource: azuread_application_identifier_uris

Manages the Identifier URI(s) of an application with Azure Active Directory.

## API Permissions

The following API permissions are required in order to use this resource.

When authenticated with a service principal, this resource requires one of the following application roles: `Application.ReadWrite.All` or `Directory.ReadWrite.All`

-> It's possible to use this resource with the `Application.ReadWrite.OwnedBy` application role, provided the principal being used to run Terraform is included in the `owners` property.

When authenticated with a user principal, this resource requires one of the following directory roles: `Application Administrator` or `Global Administrator`

## Example Usage

*Basic example*

```terraform
resource "azuread_application" "example" {
  display_name = "example"

  lifecycle {
    ignore_changes = [
      identifier_uris,
    ]
  }
}

resource "azuread_application_identifier_uris" "example" {
  application_object_id = azuread_application.example.object_id
  identifier_uris       = ["api://${azuread_application.example.application_id}"]
}
```

*Example using default onmicrosoft domain*

```terraform
resource "azuread_application" "example" {
  display_name = "example"

  lifecycle {
    ignore_changes = [
      identifier_uris,
    ]
  }
}

data "azuread_domains" "aad_domains" {
  only_default = true
}

resource "azuread_application_identifier_uris" "example" {
  application_object_id = azuread_application.example.object_id
  identifier_uris       = ["api://${azuread_application.example.application_id}", "https://exampleapp.${data.azuread_domains.aad_domains.domains.0.domain_name}"]
}
```

## Argument Reference

The following arguments are supported:

* `application_object_id` (Required) The object ID of the application for which the identifier_uris should be managed. Changing this field forces a new resource to be created.
* `identifier_uris` (Required) The user-defined URI(s) that uniquely identify an application within its Azure AD tenant, or within a verified custom domain if the application is multi-tenant
