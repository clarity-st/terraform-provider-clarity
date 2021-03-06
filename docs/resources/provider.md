---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "clarity_provider Resource - terraform-provider-clarity"
subcategory: ""
description: |-
  A Clarity provider.
---

# clarity_provider (Resource)

A Clarity [provider](https://docs.clarity.st/deployment/providers.html).


<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `name` (String) A name for the provider.

### Optional

- `aws` (Block Set, Max: 1) AWS Provider configuration. (see [below for nested schema](#nestedblock--aws))
- `slug` (String) A slug for this provider.
- `webhook` (Block Set, Max: 1) Webhook configuration. (see [below for nested schema](#nestedblock--webhook))

### Read-Only

- `id` (String) The ID of this resource.

<a id="nestedblock--aws"></a>
### Nested Schema for `aws`

Required:

- `account_id` (String) AWS Account ID
- `region` (String) AWS Region
- `role` (String) IAM Role name

Optional:

- `additional_account_id` (String) Additional AWS Account ID


<a id="nestedblock--webhook"></a>
### Nested Schema for `webhook`

Required:

- `url` (String) URL
