---
subcategory: "Object Storage Service (OBS)"
layout: "opentelekomcloud"
page_title: "OpenTelekomCloud: opentelekomcloud_obs_bucket"
sidebar_current: "docs-opentelekomcloud-resource-obs-bucket"
description: |-
  Manages a OBS Bucket resource within OpenTelekomCloud.
---

Up-to-date reference of API arguments for OBS bucket you can get at
[documentation portal](https://docs.otc.t-systems.com/object-storage-service/api-ref/apis/operations_on_buckets)

# opentelekomcloud_obs_bucket

Provides an OBS bucket resource within OpenTelekomCloud.
Now respects HTTP_PROXY, HTTPS_PROXY environment variables.

## Example Usage

### Private Bucket with Tags

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  tags = {
    foo = "bar"
    Env = "Test"
  }
}
```

### Parallel file system bucket

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket      = "my-tf-test-bucket"
  parallel_fs = true
}
```

### Enable versioning

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket     = "my-tf-test-bucket"
  acl        = "private"
  versioning = true
}
```

### Suspend versioning

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket     = "my-tf-test-bucket"
  acl        = "private"
  versioning = false
}
```

### WORM policy with versioning enabled

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket     = "my-tf-test-bucket"
  versioning = true
  worm_policy {
    years = 1
  }
}
```

### Enable Logging

```hcl
resource "opentelekomcloud_obs_bucket" "log_bucket" {
  bucket = "my-tf-log-bucket"
  acl    = "log-delivery-write"
}

resource "opentelekomcloud_obs_bucket" "b" {
  bucket = "my-tf-test-bucket"
  acl    = "private"

  logging {
    target_bucket = opentelekomcloud_obs_bucket.log_bucket.id
    target_prefix = "log/"
  }
}
```

### Static Website Hosting

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket = "obs-website-test.hashicorp.com"
  acl    = "public-read"

  website {
    index_document = "index.html"
    error_document = "error.html"

    routing_rules = <<EOF
[{
    "Condition": {
        "KeyPrefixEquals": "docs/"
    },
    "Redirect": {
        "ReplaceKeyPrefixWith": "documents/"
    }
}]
EOF
  }
}
```

### Using CORS

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket = "obs-website-test.hashicorp.com"
  acl    = "public-read"

  cors_rule {
    allowed_origins = ["https://obs-website-test.hashicorp.com"]
    allowed_methods = ["PUT", "POST"]
    allowed_headers = ["*"]
    expose_headers  = ["ETag"]
    max_age_seconds = 3000
  }
}
```

### Using server side encryption for the bucket

```hcl
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket        = "my-bucket"
  storage_class = "WARM"
  acl           = "public-read"

  server_side_encryption {
    algorithm  = "kms"
    kms_key_id = var.kms_master_key_id
  }
}
```

### Using server side encryption with the least amount of required KMS privileges

```hcl
resource "opentelekomcloud_identity_role_v3" "this_kms" {
  description   = var.role_kms_desc
  display_name  = var.role_kms_display_name
  display_layer = "project"

  statement {
    effect = "Allow"
    action = [
      "kms:cmk:list",
      "kms:cmk:get"
    ]
  }
  statement {
    effect = "Allow"
    resource = [
      "KMS:*:*:KeyId:${opentelekomcloud_kms_key_v1.this.id}"
    ]
    action = [
      "kms:cmk:generate",
      "kms:dek:create",
      "kms:cmk:crypto",
      "kms:dek:crypto"
    ]
  }

}

resource "opentelekomcloud_identity_role_v3" "this_obs" {
  description   = var.role_obs_desc
  display_name  = var.role_obs_display_name
  display_layer = "domain"
  statement {
    effect = "Allow"
    resource = [
      "OBS:*:*:bucket:${opentelekomcloud_obs_bucket.this.id}",
      "OBS:*:*:object:*"
    ]
    action = [
      "obs:object:DeleteObject",
      "obs:object:PutObject",
      "obs:object:GetObject",
      "obs:bucket:ListBucket",
      "obs:bucket:GetEncryptionConfiguration"
    ]
  }
}

resource "opentelekomcloud_identity_user_v3" "this" {
  name        = var.user_name
  description = var.user_desc
  access_type = "programmatic"

  lifecycle {
    ignore_changes = [pwd_reset]
  }
}

resource "opentelekomcloud_identity_group_v3" "this" {
  name        = var.user_group_name
  description = var.user_group_desc
}

resource "opentelekomcloud_identity_role_assignment_v3" "this_kms" {
  group_id     = opentelekomcloud_identity_group_v3.this.id
  domain_id    = var.domain_id
  role_id      = opentelekomcloud_identity_role_v3.this_kms.id
  all_projects = true
}

resource "opentelekomcloud_identity_role_assignment_v3" "this_obs" {
  group_id     = opentelekomcloud_identity_group_v3.this.id
  domain_id    = var.domain_id
  role_id      = opentelekomcloud_identity_role_v3.this_obs.id
  all_projects = true
}

resource "opentelekomcloud_kms_key_v1" "this" {
  key_alias       = var.kms_alias
  pending_days    = "7"
  key_description = var.kms_desc
  realm           = var.region
  is_enabled      = true

}

resource "opentelekomcloud_obs_bucket" "this" {
  bucket = var.bucket_name
  acl    = "private"
  server_side_encryption {
    algorithm  = "kms"
    kms_key_id = opentelekomcloud_kms_key_v1.this.id
  }
}

```

### Using object lifecycle

```hcl
resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket     = "my-bucket"
  acl        = "private"
  versioning = true

  lifecycle_rule {
    name    = "log"
    prefix  = "log/"
    enabled = true

    expiration {
      days = 365
    }
    transition {
      days          = 60
      storage_class = "WARM"
    }
    transition {
      days          = 180
      storage_class = "COLD"
    }
  }

  lifecycle_rule {
    name    = "tmp"
    prefix  = "tmp/"
    enabled = true

    noncurrent_version_expiration {
      days = 180
    }
    noncurrent_version_transition {
      days          = 30
      storage_class = "WARM"
    }
    noncurrent_version_transition {
      days          = 60
      storage_class = "COLD"
    }
  }
}
```

### Using event notifications

```hcl
resource "opentelekomcloud_smn_topic_v2" "topic" {
  name         = "obs-notifications"
  display_name = "The display name of topic_1"
}

resource "opentelekomcloud_smn_topic_attribute_v2" "policy" {
  topic_urn       = opentelekomcloud_smn_topic_v2.topic.id
  attribute_name  = "access_policy"
  topic_attribute = <<EOF
{
  "Version": "2016-09-07",
  "Id": "__default_policy_ID",
  "Statement": [
    {
      "Sid": "__service_pub_0",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "obs",
          "s3"
        ]
      },
      "Action": [
        "SMN:Publish",
        "SMN:QueryTopicDetail"
      ],
      "Resource": "${opentelekomcloud_smn_topic_v2.topic.id}"
    }
  ]
}
EOF

}

resource "opentelekomcloud_obs_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"

  event_notifications {
    topic = opentelekomcloud_smn_topic_v2.topic.id
    events = [
      "ObjectCreated:*",
      "ObjectRemoved:*",
    ]
    filter_rule {
      name  = "prefix"
      value = "smn"
    }
    filter_rule {
      name  = "suffix"
      value = ".jpg"
    }
  }

  depends_on = [opentelekomcloud_smn_topic_attribute_v2.policy]
}
```

### Bucket with set user domain names

```hcl
resource "opentelekomcloud_obs_bucket" "b" {
  bucket = "my-tf-test-bucket"
  user_domain_names = [
    var.domain_1,
    var.domain_2,
    var.domain_3
  ]
}
```

## Argument Reference

The following arguments are supported:

* `bucket` - (Required) Specifies the name of the bucket. Changing this parameter will create a new resource.
  A bucket must be named according to the globally applied DNS naming regulations as follows:
  * The name must be globally unique in OBS.
  * The name must contain 3 to 63 characters. Only lowercase letters, digits, hyphens (-), and periods (.) are allowed.
  * The name cannot start or end with a period (.) or hyphen (-), and cannot contain two consecutive periods (.)
    or contain a period (.) and a hyphen (-) adjacent to each other.
  * The name cannot be an IP address.
  * If the name contains any periods (.), a security certificate verification message may appear when you access
    the bucket or its objects by entering a domain name.

* `storage_class` - (Optional) Specifies the storage class of the bucket. OBS provides three storage classes:
  `STANDARD`, `WARM` (Infrequent Access) and `COLD` (Archive). Defaults to `STANDARD`.

* `parallel_fs` - (Optional) Whether enable a bucket as a parallel file system.

* `acl` - (Optional) Specifies the ACL policy for a bucket. The predefined common policies are as follows:
  `private`, `public-read`, `public-read-write` and `log-delivery-write`. Defaults to `private`.

* `tags` - (Optional) A mapping of tags to assign to the bucket. Each tag is represented by one key-value pair.

* `versioning` - (Optional) Set to `true` to enable versioning. Once you version-enable a bucket, it can never return to an
  unversioned state. You can, however, suspend versioning on that bucket. If omitted, during bucket
  creation it will be in `Disabled` state.

* `logging` - (Optional) A settings of bucket logging (documented below).

* `worm_policy` - (Optional) A settings of bucket default WORM policy and a retention period (documented below).
  `worm_policy` requires `versioning` to be enabled.

* `website` - (Optional) A website object (documented below).

* `cors_rule` - (Optional) A rule of Cross-Origin Resource Sharing (documented below).

* `lifecycle_rule` - (Optional) A configuration of object lifecycle management (documented below).

* `server_side_encryption` - (Optional) A configuration of server side encryption (documented below).

* `event_notifications` - (Optional) A configuration of bucket event notifications (documented below).

* `force_destroy` - (Optional) A boolean that indicates all objects should be deleted from the bucket so that the
  bucket can be destroyed without error. Default to `false`.

* `region` - (Optional) If specified, the region this bucket should reside in. Otherwise,
  the region used by the provider.

* `user_domain_names` - (Optional) Specifies the user domain names. The restriction requirements for this field
  are as follows:
    + Each value must meet the domain name rules.
    + The maximum length of a domain name is 256 characters.
    + A custom domain name can only be used by one bucket.

  -> When creating or updating the OBS bucket user domain names, the original user domain names will be overwritten.

The `logging` object supports the following:

* `target_bucket` - (Required) The name of the bucket that will receive the log objects.
  The acl policy of the target bucket should be `log-delivery-write`.

* `target_prefix` - (Optional) To specify a key prefix for log objects.

* `agency` - (Optional) Specifies the IAM agency of OBS cloud service.

  -> The IAM agency requires the `PutObject` permission for the target bucket.  If default encryption is enabled for the
  target bucket, the agency also requires the `KMS Administrator` permission in the region where the target bucket is
  located.

The `worm_policy` object supports the following:

* `days` - (Optional) Default protection period, in `days`.
  The value is from `1` to `36500`.

* `years` - (Optional) Default protection period, in years. In a leap year, only 365 days are calculated.
  The value is from `1` to `100`.

The `website` object supports the following:

* `index_document` - (Required, unless using `redirect_all_requests_to`) Specifies the default homepage of the
  static website, only HTML web pages are supported. OBS only allows files such as `index.html` in the root
  directory of a bucket to function as the default homepage. That is to say, do not set the default homepage
  with a multi-level directory structure (for example, `/page/index.html`).

* `error_document` - (Optional) Specifies the error page returned when an error occurs during static website access.
  Only HTML, JPG, PNG, BMP, and WEBP files under the root directory are supported.

* `redirect_all_requests_to` - (Optional) A hostname to redirect all website requests for this bucket to.
  Hostname can optionally be prefixed with a protocol (`http://` or `https://`) to use when redirecting
  requests. The default is the protocol that is used in the original request.

* `routing_rules` - (Optional) A JSON or XML format containing routing rules describing redirect
  behavior and when redirects are applied. Each rule contains a `Condition` and a `Redirect`
  as shown in the following table:

| Parameter | Key |
|-----------|-----|
| Condition | KeyPrefixEquals, HttpErrorCodeReturnedEquals |
| Redirect | Protocol, HostName, ReplaceKeyPrefixWith, ReplaceKeyWith, HttpRedirectCode |

The `cors_rule` object supports the following:

* `allowed_origins` - (Required) Requests from this origin can access the bucket. Multiple matching rules are allowed.
  One rule occupies one line, and allows one wildcard character (*) at most.

* `allowed_methods` - (Required) Specifies the acceptable operation type of buckets and objects.
  The methods include `GET`, `PUT`, `POST`, `DELETE` or `HEAD`.

* `allowed_headers` - (Optional) Specifies the allowed header of cross-origin requests.
  Only CORS requests matching the allowed header are valid.

* `expose_headers` - (Optional) Specifies the exposed header in CORS responses, providing additional
  information for clients.

* `max_age_seconds` - (Optional) Specifies the duration that your browser can cache CORS responses,
  expressed in seconds. The default value is `100`.

The `lifecycle_rule` object supports the following:

* `name` - (Required) Unique identifier for lifecycle rules. The Rule Name contains a maximum of 255 characters.

* `enabled` - (Required) Specifies lifecycle rule status.

* `prefix` - (Optional) Object key prefix identifying one or more objects to which the rule applies.
  If omitted, all objects in the bucket will be managed by the lifecycle rule. The prefix cannot start
  or end with a slash (/), cannot have consecutive slashes (/), and cannot contain the following
  special characters: \:*?"<>|.

* `expiration` - (Optional) Specifies a period when objects that have been last updated are automatically
  deleted. (documented below).

* `transition` - (Optional) Specifies a period when objects that have been last updated are automatically
  transitioned to `WARM` or `COLD` storage class (documented below).

* `noncurrent_version_expiration` - (Optional) Specifies a period when noncurrent object versions are
  automatically deleted. (documented below).

* `noncurrent_version_transition` - (Optional) Specifies a period when noncurrent object versions are
  automatically transitioned to `WARM` or `COLD` storage class (documented below).

-> At least one of `expiration`, `transition`, `noncurrent_version_expiration`, `noncurrent_version_transition`
must be specified.

The `expiration` object supports the following

* `days` - (Required) Specifies the number of days when objects that have been last updated are automatically deleted.
  The expiration time must be greater than the transition times.

The `transition` object supports the following

* `days` - (Required) Specifies the number of days when objects that have been last updated are automatically
  transitioned to the specified storage class.

* `storage_class` - (Required) The class of storage used to store the object. Only `WARM` and `COLD` are supported.

The `noncurrent_version_expiration` object supports the following

* `days` - (Required) Specifies the number of days when noncurrent object versions are automatically deleted.

The `noncurrent_version_transition` object supports the following

* `days` - (Required) Specifies the number of days when noncurrent object versions are automatically
  transitioned to the specified storage class.

* `storage_class` - (Required) The class of storage used to store the object. Only `WARM` and `COLD` are supported.

The `server_side_encryption` object supports the following

* `algorithm` - (Required) The algorithm used for SSE. Only `kms` is supported.

* `kms_key_id` - (Required) The ID of KMS key used for the encryption.

~> Only base project (e.g. `eu-de`) KMS keys can be used for the encryption

The `event_notifications` object supports the following

* `id` - (Optional) Unique ID of the event notification. If the user does not specify an ID, the system assigns an ID automatically.

* `topic` - (Required) URN of the event notification topic. After detecting a specific event, OBS sends a message to the topic.

-> Topic should exist and be authorized to be used by OBS.

* `events` - (Required) Type of events that need to be notified.

-> Events should not have `s3:` prefix, e.g. `"ObjectCreated:*"` is valid value, but `"s3:ObjectCreated:*"` is not.

* `filter_rule` - (Optional) Filtering rules. The rules filter objects based on the prefixes and suffixes of object names.

The `filter_rule` object supports the following

* `name` - (Optional) Specifies the prefix or suffix of object names for filtering. Valid values are `prefix` and `suffix`.

* `value` - (Optional) Specifies keywords of object names so that objects can be filtered based on the prefixes or suffixes.
  The value contains a maximum of `1024` characters.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - The name of the bucket.

* `bucket_domain_name` - The bucket domain name. Will be of format `bucketname.obs.region.otc.t-systems.com`.

* `region` - The region this bucket resides in.

* `bucket_version` - The OBS version of the bucket.

## Import

OBS bucket can be imported using the `bucket`, e.g.

```shell
terraform import opentelekomcloud_obs_bucket.bucket bucket-name
```
