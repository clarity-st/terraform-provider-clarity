package internal

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResource(t *testing.T) {
	config := testAccProvider() + testAccService + testAccResource
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"clarity_provider.test", "slug", regexp.MustCompile("^terraform-test")),
					resource.TestMatchResourceAttr(
						"clarity_service.test", "slug", regexp.MustCompile("^terraform-test")),
					resource.TestCheckResourceAttr("clarity_resource.test", "name", "terraform-test"),
					resource.TestCheckResourceAttr("clarity_resource.test", "lambda.0.function_name", "terraform-test"),
					resource.TestCheckResourceAttr("clarity_resource.test", "lambda.0.alias", "clarity"),
					resource.TestMatchResourceAttr(
						"clarity_resource.test", "slug", regexp.MustCompile("^terraform-test")),
				),
			},
			{
				ResourceName:            "clarity_resource.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

const testAccResource = `
resource "clarity_resource" "test" {
  provider_slug = clarity_provider.test.slug
  service_slug = clarity_service.test.slug
  name = "terraform-test"

  lambda {
    function_name = "terraform-test"
  }
}
`
