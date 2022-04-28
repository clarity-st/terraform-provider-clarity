terraform {
  required_version = ">=0.13"
}

provider "clarity" {}

resource "clarity_service" "test" {
  provider_slug = "staaging"
  name          = "terraform resource"

}

resource "clarity_resource" "dev" {
  provider_slug = "staaging"
  service_slug  = clarity_service.test.slug
  name          = "dev"

  lambda {
    function_name = "api"
    alias         = "test"
  }

  deployment {
    trigger {
      manual_user_interface = false
    }
  }
}

output "resource_slug" {
  value = clarity_resource.dev.slug
}

output "service_slug" {
  value = clarity_service.test.slug
}
