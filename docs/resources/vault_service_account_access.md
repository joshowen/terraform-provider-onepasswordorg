---
page_title: "onepasswordorg_vault_service_account_access Resource - terraform-provider-onepasswordorg"
subcategory: ""
description: |-
  Provides vault access for a service account.
---

# onepasswordorg_vault_service_account_access (Resource)

Grants a service account access to a 1Password vault with a specified set of permissions.

This resource maps to the `op vault service-account grant` CLI command and uses
`op vault service-account revoke` on deletion.

## Example Usage

```terraform
resource "onepasswordorg_service_account" "example" {
  name = "example-service-account"
}

data "onepasswordorg_vault" "example" {
  name = "Example Vault"
}

resource "onepasswordorg_vault_service_account_access" "example" {
  vault_id           = data.onepasswordorg_vault.example.id
  service_account_id = onepasswordorg_service_account.example.id

  permissions {
    allow_viewing          = true
    allow_editing          = false
    allow_managing         = false
    view_items             = true
    create_items           = false
    edit_items             = false
    archive_items          = false
    delete_items           = false
    view_and_copy_passwords = true
    view_item_history      = true
    import_items           = false
    export_items           = false
    copy_and_share_items   = false
    print_items            = false
    manage_vault           = false
  }
}
```

## Schema

### Required

- `vault_id` (String) The vault ID. Changing this forces a new resource.
- `service_account_id` (String) The service account ID. Changing this forces a new resource.
- `permissions` (Block, Required) The permissions granted to the service account for this vault (see [below for nested schema](#nestedblock--permissions)).

### Read-Only

- `id` (String) The resource ID in the format `<vault_id>/<service_account_id>`.

<a id="nestedblock--permissions"></a>
### Nested Schema for `permissions`

Required:

- `allow_viewing` (Boolean)
- `allow_editing` (Boolean)
- `allow_managing` (Boolean)
- `view_items` (Boolean)
- `create_items` (Boolean)
- `edit_items` (Boolean)
- `archive_items` (Boolean)
- `delete_items` (Boolean)
- `view_and_copy_passwords` (Boolean)
- `view_item_history` (Boolean)
- `import_items` (Boolean)
- `export_items` (Boolean)
- `copy_and_share_items` (Boolean)
- `print_items` (Boolean)
- `manage_vault` (Boolean)

## Import

Import using the resource ID format `<vault_id>/<service_account_id>`:

```shell
terraform import onepasswordorg_vault_service_account_access.example <vault_id>/<service_account_id>
```
