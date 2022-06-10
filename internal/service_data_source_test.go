package internal

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccServiceDataSource(t *testing.T) {
	dataSourceName := "data.clarity_service.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProvider() + testAccService + testAccServiceDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "slug", regexp.MustCompile("^terraform-test")),
					resource.TestCheckResourceAttr(dataSourceName, "name", "terraform-test"),
				),
			},
		},
	})
}

const testAccServiceDataSource = `
data "clarity_service" "test" {
  name = clarity_service.test.name
}
`
