package internal

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProviderDataSource(t *testing.T) {
	dataSourceName := "data.clarity_provider.test"

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProvider() + testAccProviderDataSource,
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(dataSourceName, "slug", regexp.MustCompile("^terraform-test")),
					resource.TestCheckResourceAttr(dataSourceName, "name", "terraform-test"),
				),
			},
		},
	})
}

const testAccProviderDataSource = `
data "clarity_provider" "test" {
  name = clarity_provider.test.name
}
`
