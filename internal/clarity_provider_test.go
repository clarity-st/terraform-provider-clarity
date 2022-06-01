package internal

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProvider(t *testing.T) {
	account, region, role := loadAWSSettigns()
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t) },
		ProviderFactories: providerFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderWith(account, region, role),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("clarity_provider.test", "name", "terraform-test"),
					resource.TestCheckResourceAttr("clarity_provider.test", "aws.0.account_id", account),
					resource.TestCheckResourceAttr("clarity_provider.test", "aws.0.region", region),
					resource.TestCheckResourceAttr("clarity_provider.test", "aws.0.role", role),
					resource.TestMatchResourceAttr(
						"clarity_provider.test", "slug", regexp.MustCompile("^terraform-test")),
				),
			},
			{
				ResourceName:            "clarity_provider.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"force_destroy"},
			},
		},
	})
}

func testAccProvider() string {
	account, region, role := loadAWSSettigns()
	return testAccProviderWith(account, region, role)
}

func loadAWSSettigns() (string, string, string) {
	var account, role, region string
	if v, ok := os.LookupEnv("AWS_ACCOUNT_ID"); ok {
		account = v
	} else {
		account = "993614041743"
	}3
	if v, ok := os.LookupEnv("AWS_REGION"); ok {
		region = v
	} else {
		region = "us-east-1"
	}
	if v, ok := os.LookupEnv("CLARITY_PROVIDER_ROLE"); ok {
		role = v
	} else {
		role = "clarity-provider-us-east-1-staging-environment"
	}

	return account, region, role

}

func testAccProviderWith(account, region, role string) string {
	return fmt.Sprintf(`
resource "clarity_provider" "test" {
  name = "terraform-test"

  aws {
    account_id = "%s"
    region = "%s"
    role = "%s"
  }
}
`, account, region, role)
}
