package internal

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("clarity_resource.test", "provider_slug", "staaging"),
					resource.TestCheckResourceAttr("clarity_resource.test", "service_slug", "fake-event"),
					resource.TestCheckResourceAttr("clarity_resource.test", "name", "bar"),
					resource.TestCheckResourceAttr("clarity_resource.test", "lambda.0.function_name", "api"),
					resource.TestCheckResourceAttr("clarity_resource.test", "lambda.0.alias", "clarity"),
					resource.TestMatchResourceAttr(
						"clarity_resource.test", "slug", regexp.MustCompile("^ba")),
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
  provider_slug = "staaging"
  service_slug = "fake-event"
  name = "bar"

  lambda {
    function_name = "api"
  }
}
`
