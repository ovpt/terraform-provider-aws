package aws

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/schema"
)

func dataSourceAwsSubnetIDs() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceAwsSubnetIDsRead,
		Schema: map[string]*schema.Schema{

			"tags": tagsSchemaComputed(),

			"vpc_id": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"ids": &schema.Schema{
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},

			"cidr_block": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"map_public_ip_on_launch": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"default_for_az": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
			},

			"availability_zone": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func dataSourceAwsSubnetIDsRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).ec2conn

	req := &ec2.DescribeSubnetsInput{}

	defaultForAzStr := ""
	if d.Get("default_for_az").(bool) {
		defaultForAzStr = "true"
	}

	mapPublicIpOnLaunchStr := ""
	if d.Get("map_public_ip_on_launch").(bool) {
		mapPublicIpOnLaunchStr = "false"
	}

	req.Filters = buildEC2AttributeFilterList(
		map[string]string{
			"vpc-id":              d.Get("vpc_id").(string),
			"cidrBlock":           d.Get("cidr_block").(string),
			"mapPublicIpOnLaunch": mapPublicIpOnLaunchStr,
			"defaultForAz":        defaultForAzStr,
			"availabilityZone":    d.Get("availability_zone").(string),
			"state":               d.Get("state").(string),
		},
	)

	req.Filters = append(req.Filters, buildEC2TagFilterList(
		tagsFromMap(d.Get("tags").(map[string]interface{})),
	)...)

	log.Printf("[DEBUG] DescribeSubnets %s\n", req)
	resp, err := conn.DescribeSubnets(req)
	if err != nil {
		return err
	}

	if resp == nil || len(resp.Subnets) == 0 {
		return fmt.Errorf("no matching subnet found for vpc with id %s", d.Get("vpc_id").(string))
	}

	subnets := make([]string, 0)

	for _, subnet := range resp.Subnets {
		subnets = append(subnets, *subnet.SubnetId)
	}

	d.SetId(d.Get("vpc_id").(string))
	d.Set("ids", subnets)

	return nil
}
