package internal

import (
	"context"
	"fmt"

	"github.com/clarity-st/terraform-provider-clarity/internal/clarity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func serviceDatasource() *schema.Resource {
	return &schema.Resource{
		ReadContext: serviceDatasourceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name for the service.",
				Required:    true,
				Type:        schema.TypeString,
			},
			"slug": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func serviceDatasourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)
	name := d.Get("name").(string)

	rsp, err := client.ListServices()
	if err != nil {
		return diag.FromErr(err)
	}

	matches := 0
	for _, r := range rsp.Services {
		if r.Name == name {
			matches++
		}
	}
	if matches == 0 {
		return diag.FromErr(fmt.Errorf("No matching service found with the name '%s'", name))
	}

	if matches > 1 {
		return diag.FromErr(fmt.Errorf("Found multiple service with the name '%s'", name))
	}

	for _, r := range rsp.Services {
		if r.Name == name {
			d.Set("slug", r.Slug)
			d.SetId(r.Slug)
			return nil
		}
	}

	return nil
}
