---
subcategory: "Identity and Access Management (IAM)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_identity_user_group_membership_v3"
sidebar_current: "docs-opentelekomcloud-resource-identity-user-group-membership-v3"
description: |-
  Manages a IAM User Group Membership resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for IAM user group membership you can get at
[documentation portal](https://docs.otc.t-systems.com/identity-access-management/api-ref/apis/user_group_management)

# opentelekomcloud_identity_user_group_membership_v3

Manages a User Group Membership resource within OpenTelekomCloud IAM service.

-> **Note:** You _must_ have admin privileges in your OpenTelekomCloud cloud to use this resource.

## Example Usage

```hcl
resource "opentelekomcloud_identity_user_v3" "user_1" {
  name     = "user-1"
  password = "password123@#"
  enabled  = true
}

resource "opentelekomcloud_identity_group_v3" "group_1" {
  name = "group-1"
}

resource "opentelekomcloud_identity_group_v3" "group_2" {
  name = "group-2"
}

resource "opentelekomcloud_identity_user_group_membership_v3" "membership_1" {
  user = opentelekomcloud_identity_user_v3.user_1.id
  groups = [
    opentelekomcloud_identity_group_v3.group_1.id,
    opentelekomcloud_identity_group_v3.group_2.id,
  ]
}
```

## Argument Reference

The following arguments are supported:

* `user` - (Required) ID of a user.

* `groups` - (Required) IDs of the groups for the user to be assigned to.

## Attributes Reference

The following attributes are exported:

* `user` - See Argument Reference above.

* `groups` - See Argument Reference above.
