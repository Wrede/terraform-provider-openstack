package openstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/layer3/routers"
)

func TestAccNetworkingV2Router_basic(t *testing.T) {
	var router routers.Router

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2RouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2RouterBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2RouterExists("openstack_networking_router_v2.router_1", &router),
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_1", "description", "router description"),
				),
			},
			{
				Config: testAccNetworkingV2RouterUpdate,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_1", "name", "router_2"),
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_1", "description", ""),
				),
			},
		},
	})
}

func TestAccNetworkingV2Router_updateExternalGateway(t *testing.T) {
	var router routers.Router

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2RouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2RouterUpdateExternalGateway1,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2RouterExists("openstack_networking_router_v2.router_1", &router),
				),
			},
			{
				Config: testAccNetworkingV2RouterUpdateExternalGateway2,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_1", "external_network_id", osExtGwID),
				),
			},
		},
	})
}

func TestAccNetworkingV2Router_vendor_opts(t *testing.T) {
	var router routers.Router

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2RouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2RouterVendorOpts,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2RouterExists("openstack_networking_router_v2.router_1", &router),
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_1", "external_gateway", osExtGwID),
				),
			},
		},
	})
}

func TestAccNetworkingV2Router_vendor_opts_no_snat(t *testing.T) {
	var router routers.Router

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAdminOnly(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2RouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2RouterVendorOptsNoSnat,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkingV2RouterExists("openstack_networking_router_v2.router_1", &router),
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_1", "external_gateway", osExtGwID),
				),
			},
		},
	})
}

func TestAccNetworkingV2Router_extFixedIPs(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAdminOnly(t)
		},
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckNetworkingV2RouterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkingV2RouterExtFixedIPs,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_2", "name", "router_2"),
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_2", "external_fixed_ip.#", "2"),
					resource.TestCheckResourceAttr(
						"openstack_networking_router_v2.router_2", "enable_snat", "true"),
				),
			},
		},
	})
}

func testAccCheckNetworkingV2RouterDestroy(s *terraform.State) error {
	config := testAccProvider.Meta().(*Config)
	networkingClient, err := config.NetworkingV2Client(osRegionName)
	if err != nil {
		return fmt.Errorf("Error creating OpenStack networking client: %s", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "openstack_networking_router_v2" {
			continue
		}

		_, err := routers.Get(networkingClient, rs.Primary.ID).Extract()
		if err == nil {
			return fmt.Errorf("Router still exists")
		}
	}

	return nil
}

func testAccCheckNetworkingV2RouterExists(n string, router *routers.Router) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		config := testAccProvider.Meta().(*Config)
		networkingClient, err := config.NetworkingV2Client(osRegionName)
		if err != nil {
			return fmt.Errorf("Error creating OpenStack networking client: %s", err)
		}

		found, err := routers.Get(networkingClient, rs.Primary.ID).Extract()
		if err != nil {
			return err
		}

		if found.ID != rs.Primary.ID {
			return fmt.Errorf("Router not found")
		}

		*router = *found

		return nil
	}
}

const testAccNetworkingV2RouterBasic = `
resource "openstack_networking_router_v2" "router_1" {
  name = "router_1"
  description = "router description"
  admin_state_up = "true"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`

const testAccNetworkingV2RouterUpdate = `
resource "openstack_networking_router_v2" "router_1" {
  name = "router_2"
  admin_state_up = "true"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`

var testAccNetworkingV2RouterVendorOpts = fmt.Sprintf(`
resource "openstack_networking_router_v2" "router_1" {
  name = "router_1"
  admin_state_up = "true"
  external_network_id = "%s"
  vendor_options {
    set_router_gateway_after_create = true
  }
}
`, osExtGwID)

var testAccNetworkingV2RouterVendorOptsNoSnat = fmt.Sprintf(`
resource "openstack_networking_router_v2" "router_1" {
  name = "router_1"
  admin_state_up = "true"
  distributed = "false"
  external_network_id = "%s"
  enable_snat = "false"
  vendor_options {
    set_router_gateway_after_create = true
  }
}
`, osExtGwID)

const testAccNetworkingV2RouterUpdateExternalGateway1 = `
resource "openstack_networking_router_v2" "router_1" {
  name = "router"
  admin_state_up = "true"
}
`

var testAccNetworkingV2RouterUpdateExternalGateway2 = fmt.Sprintf(`
resource "openstack_networking_router_v2" "router_1" {
  name = "router"
  admin_state_up = "true"
  external_network_id = "%s"
}
`, osExtGwID)

var testAccNetworkingV2RouterExtFixedIPs = fmt.Sprintf(`
resource "openstack_networking_router_v2" "router_1" {
  name = "router_1"
  admin_state_up = "true"
  external_network_id = "%s"

  timeouts {
    create = "5m"
    delete = "5m"
  }
}

resource "openstack_networking_router_v2" "router_2" {
  name = "router_2"
  admin_state_up = "true"
  external_network_id = "%s"

  external_fixed_ip {
    subnet_id = "${openstack_networking_router_v2.router_1.external_fixed_ip.0.subnet_id}"
  }

  external_fixed_ip {
    subnet_id = "${openstack_networking_router_v2.router_1.external_fixed_ip.0.subnet_id}"
  }

  timeouts {
    create = "5m"
    delete = "5m"
  }
}
`, osExtGwID, osExtGwID)
