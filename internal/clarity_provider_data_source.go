package internal

import (
	"context"
	"fmt"

	"github.com/clarity-st/terraform-provider-clarity/internal/clarity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func providerDatasource() *schema.Resource {
	return &schema.Resource{
		ReadContext: providerDatasourceRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Description: "A name for this provider.",
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

func providerDatasourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)
	name := d.Get("name").(string)

	rsp, err := client.LoadProviders()
	if err != nil {
		return diag.FromErr(err)
	}

	matches := 0
	for _, r := range rsp {
		if r.Name == name {
			matches++
		}
	}
	if matches == 0 {
		return diag.FromErr(fmt.Errorf("No matching provider found with the name '%s'", name))
	}

	if matches > 1 {
		return diag.FromErr(fmt.Errorf("Found multiple providers with the name '%s'", name))
	}

	for _, r := range rsp {
		if r.Name == name {
			d.Set("slug", r.Slug)
			d.SetId(r.Slug)
			return nil
		}
	}

	return nil
}
