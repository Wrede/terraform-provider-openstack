package openstack

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccFWPolicyV1_importBasic(t *testing.T) {
	resourceName := "openstack_fw_policy_v1.policy_1"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckFW(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFWPolicyV1Destroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFWPolicyV1AddRules,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
