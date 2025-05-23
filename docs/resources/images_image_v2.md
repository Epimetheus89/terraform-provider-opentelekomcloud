---
subcategory: "Image Management Service (IMS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_images_image_v2"
sidebar_current: "docs-opentelekomcloud-resource-images-image-v2"
description: |-
  Manages an Image Management resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for Image management you can get at
[documentation portal](https://docs.otc.t-systems.com/image-management-service/api-ref/native_openstack_apis/image_native_openstack_apis)

# opentelekomcloud_images_image_v2

Manages a V2 Image resource within OpenTelekomCloud Glance.

## Example Usage

```hcl
resource "opentelekomcloud_images_image_v2" "rancheros" {
  name             = "RancherOS"
  image_source_url = "https://releases.rancher.com/os/latest/rancheros-openstack.img"
  container_format = "bare"
  disk_format      = "qcow2"
  hw_firmware_type = "uefi"

  tags = ["foo.bar", "tag.value"]
}
```

## Argument Reference

The following arguments are supported:

* `container_format` - (Required, String, ForceNew) The container format. Must be one of
  `ami`, `ari`, `aki`, `bare`, `ovf`.

* `disk_format` - (Required, String, ForceNew) The disk format. Must be one of
  `ami`, `ari`, `aki`, `vhd`, `vmdk`, `raw`, `qcow2`, `vdi`, `iso`.

* `local_file_path` - (Optional, String, ForceNew) This is the filepath of the raw image file
  that will be uploaded to Glance. Conflicts with `image_source_url`.

* `image_cache_path` - (Optional, String) This is the directory where the images will
  be downloaded. Images will be stored with a filename corresponding to
  the url's md5 hash. Defaults to "$HOME/.terraform/image_cache"

* `image_source_url` - (Optional, String, ForceNew) This is the url of the raw image that will
  be downloaded in the `image_cache_path` before being uploaded to Glance.
  Glance is able to download image from internet but the `gophercloud` library
  does not yet provide a way to do so.
  Conflicts with `local_file_path`.

* `min_disk_gb` - (Optional, Integer, ForceNew) Amount of disk space (in GB) required to boot image.
  Defaults to 0.

* `min_ram_mb` - (Optional, Integer, ForceNew) Amount of ram (in MB) required to boot image.
  Defauts to 0.

* `name` - (Required, String) The name of the image.

* `protected` - (Optional, Boolean, ForceNew) If true, image will not be deletable.
  Defaults to false.

* `tags` - (Optional, List) The tags of the image. It must be a list of strings.
  At this time, it is not possible to delete all tags of an image.

* `visibility` - (Optional, String) The visibility of the image. Must be one of
  "public", "private", "community", or "shared". The ability to set the
  visibility depends upon the configuration of the OpenTelekomCloud cloud.

* `hw_firmware_type` - (Optional, String) Specifies the boot mode. The value can be `bios` or `uefi`.

-> **Note:** The `properties` attribute handling in the gophercloud library is currently buggy
and needs to be fixed before being implemented in this resource.

## Attributes Reference

In additin to the arguments defined above, the following attributes are exported:

* `checksum` - The checksum of the data associated with the image.

* `created_at` - The date the image was created.


* `file` - the trailing path after the glance
  endpoint that represent the location of the image
  or the path to retrieve it.

* `id` - A unique ID assigned by Glance.

* `owner` - The id of the opentelekomcloud user who owns the image.

* `schema` - The path to the JSON-schema that represent
  the image or image

* `size_bytes` - The size in bytes of the data associated with the image.

* `status` - The status of the image. It can be `queued`, `active`
  or `saving`.

* `update_at` - The date the image was last updated.

## Import

Images can be imported using the `id`, e.g.

```sh
terraform import opentelekomcloud_images_image_v2.rancheros 89c60255-9bd6-460c-822a-e2b959ede9d2
```
