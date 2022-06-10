terraform {
  required_version = ">=0.13"

  required_providers {
    clarity = {
      source = "local/clarity"
    }
  }
}

provider "clarity" {}

data "clarity_provider" "test" {
  name = "Terraform data test"
}

resource "clarity_service" "test" {
  provider_slug = data.clarity_provider.test.slug
  name          = "terraform resource"
}

data "clarity_service" "test" {
  name = "terraform data test"
}

resource "clarity_resource" "test" {
  service_slug = data.clarity_service.test.slug
  name         = "terraform resource"
}
