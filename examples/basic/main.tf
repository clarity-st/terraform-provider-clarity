terraform {
  required_version = ">=0.13"

  required_providers {
    clarity = {
      source = "local/clarity"
    }
  }
}

provider "clarity" {}

resource "clarity_provider" "test" {
  name = "test provider"

  aws {
    account_id = "993614041743"
    region     = "us-east-1"
    role       = "clarity-provider-us-east-1-staging-environment"
  }
}

resource "clarity_service" "test" {
  provider_slug = clarity_provider.test.slug
  name          = "terraform resource"
}

resource "clarity_resource" "dev" {
  provider_slug = clarity_provider.test.slug
  service_slug  = clarity_service.test.slug
  name          = "dev"

  lambda {
    function_name = "terraform-test"
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
