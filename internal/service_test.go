package internal

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccService(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccService,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("clarity_service.test", "provider_slug", "staaging"),
					resource.TestCheckResourceAttr("clarity_service.test", "name", "bar"),
					resource.TestMatchResourceAttr(
						"clarity_service.test", "slug", regexp.MustCompile("^ba")),
				),
			},
			{
				ResourceName:            "clarity_service.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

const testAccService = `
resource "clarity_service" "test" {
  provider_slug = "staaging"
  name = "bar"
}
`
