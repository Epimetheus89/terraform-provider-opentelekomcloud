---
subcategory: "Cloud Backup and Recovery (CBR)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_cbr_vault_v3"
sidebar_current: "docs-opentelekomcloud-resource-cbr-vault-v3"
description: |-
  Manages a CBR Vault resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for CBR vault you can get at
[documentation portal](https://docs.otc.t-systems.com/cloud-backup-recovery/api-ref/cbr_apis/vaults)

# opentelekomcloud_cbr_vault_v3

Manages a V3 CBR Vault resource within OpenTelekomCloud.

## Example usage

### Simple vault

```hcl
resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for terraform provider test"

  billing {
    size          = 100
    object_type   = "disk"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }
}
```

### Vault with associated resource (server)

```hcl
resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for terraform provider test"

  billing {
    size          = 100
    object_type   = "disk"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }

  resource {
    id   = opentelekomcloud_ecs_instance_v1.instance.id
    type = "OS::Nova::Server"

    exclude_volumes = [
      opentelekomcloud_ecs_instance_v1.instance_1.data_disks.1.id
    ]
  }
}
```
Include volumes works currently only on SwissCloud:
```hcl
resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for terraform provider test"

  billing {
    size          = 100
    object_type   = "disk"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }

  resource {
    id   = opentelekomcloud_ecs_instance_v1.instance.id
    type = "OS::Nova::Server"

    include_volumes = [
      opentelekomcloud_ecs_instance_v1.instance_1.data_disks.1.id
    ]
  }
}
```

### Vault with associated resource (volume)

```hcl
resource "opentelekomcloud_blockstorage_volume_v2" "volume" {
  name = "cbr-test-volume"
  size = 10

  volume_type = "SSD"
}

resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for terraform provider test"

  billing {
    size          = 100
    object_type   = "disk"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }

  resource {
    id   = opentelekomcloud_blockstorage_volume_v2.volume.id
    type = "OS::Cinder::Volume"
  }
}
```

### Vault with associated resource (sfs-turbo)

```hcl
variable "vpc_id" {}
variable "subnet_id" {}
variable "sg_id" {}
variable "az" {}

resource "opentelekomcloud_sfs_turbo_share_v1" "sfs-turbo" {
  name              = "sfs-turbo-share"
  size              = 500
  share_proto       = "NFS"
  vpc_id            = var.vpc_id
  subnet_id         = var.subnet_id
  security_group_id = var.sg_id
  availability_zone = var.az
}

resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for terraform provider test"

  billing {
    size          = 1000
    object_type   = "turbo"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }

  resource {
    id   = opentelekomcloud_sfs_turbo_share_v1.sfs-turbo.id
    type = "OS::Sfs::Turbo"
  }
}
```

### Vault with associated policy

```hcl
resource "opentelekomcloud_cbr_policy_v3" "policy" {
  name           = "some-policy"
  operation_type = "backup"

  trigger_pattern = [
    "FREQ=WEEKLY;BYDAY=MO,TU,WE,TH,FR,SA,SU;BYHOUR=14;BYMINUTE=00"
  ]
  operation_definition {
    day_backups   = 1
    week_backups  = 2
    year_backups  = 3
    month_backups = 4
    max_backups   = 10
    timezone      = "UTC+03:00"
  }

  enabled = "false"
}

resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for terraform provider test"

  backup_policy_id = opentelekomcloud_cbr_policy_v3.policy.id

  billing {
    size          = 100
    object_type   = "disk"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }
}
```

### Vault with auto bind and bind rule

```hcl
resource "opentelekomcloud_cbr_vault_v3" "vault" {
  name = "cbr-vault-test"

  description = "CBR vault for default backup policy"

  billing {
    size          = 10
    object_type   = "server"
    protect_type  = "backup"
    charging_mode = "post_paid"
  }
  auto_bind = true
  bind_rules {
    key   = "foo"
    value = "bar"
  }
}
```

## Argument reference

The following arguments are supported:

* `name` - Vault name.

* `billing` - Billing parameter information for creation. Billing fields:

    * `cloud_type` - (Optional) Cloud platform. One of `public` (default), `hybrid`

    * `consistent_level` - (Optional) Backup specifications. The default value is `crash_consistent`

    * `object_type` - Object type. One of `server`, `disk`, `turbo`.

    * `protect_type` - Operation type. One of `backup`, `replication`

    * `size` - Capacity, in GB. Minimum `1`, maximum `10485760`

    * `charging_mode` - Billing mode. Possible values are `post_paid` (pay-per-use) or `pre_paid`
      (yearly/monthly packages). The value defaults to `post_paid`.

    * `period_type` - (Optional) Package type. This parameter is mandatory if `charging_mode` is set to `pre_paid`.
      Possible values are `year` (yearly) or `month` (monthly).

    * `period_num` - (Optional) Required duration for the package. This parameter is mandatory if
      `charging_mode` is set to `pre_paid`.

    * `is_auto_renew` - (Optional) Whether to automatically renew the subscription after expiration. By default, it is
      not renewed.

    * `is_auto_pay` - (Optional) Whether the fee is automatically deducted from the customer's account balance after an
      order is submitted. The non-automatic payment mode is used by default.

    * `console_url` - (Optional) Redirection URL.

    * `extra_info` - (Optional) Map of extra info.

* `resource` - (Optional) Associated resources. Multiple. Resource fields:

    * `id` - ID of the resource to be backed up.

    * `type` - Type of the resource to be backed up. Possible values are `OS::Nova::Server`, `OS::Cinder::Volume` and `OS::Sfs::Turbo`.

    * `name` - (Optional) Resource name.

    * `exclude_volumes` - (Optional) List of excluded volumes.

    * `include_volumes` - (Optional) List of included volumes.

* `backup_policy_id` - (Optional) Backup policy ID. If the value of this parameter is empty, automatic backup is not
  performed.

* `description` - (Optional) User-defined vault description.

* `tags` - (Optional) Tag map.

* `auto_bind` - (Optional) Whether automatic association is supported.

* `bind_rules` - (Optional)  Tag map, a rules for automatic association. You can only select tag keys and values from
  the existing ones. If there are no tags available, go to the corresponding service to create one.
  You can add a maximum of 5 tags for a search. If more than one tag is added, the backups containing one of the
  specified tags will be returned.

* `auto_expand` - (Optional) Whether to automatically expand the vault capacity. Only pay-per-use vaults support this
  function.

## Attributes Reference

All above argument parameters can be exported as attribute parameters along with attribute reference.

* `product_id` - Product ID.

* `order_id` - Order ID.

* `allocated` - Allocated capacity, in MB.

* `spec_code` - Specification code.

* `used` - Used capacity, in MB.

* `storage_unit` - Name of the bucket for the vault.

* `frozen_scene` - Scenario when an account is frozen.

* `status` - Vault status.

## Import

Volumes can be imported using the `id`, e.g.

```sh
terraform import opentelekomcloud_cbr_vault_v3.vault ea257959-eeb1-4c10-8d33-26f0409a766b
```
