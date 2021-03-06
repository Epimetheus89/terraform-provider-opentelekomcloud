package availabilityzones

import "github.com/opentelekomcloud/gophertelekomcloud"

func listURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("os-availability-zone")
}

func listDetailURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("os-availability-zone", "detail")
}
