package internal

import (
	"context"

	"github.com/clarity-st/terraform-provider-clarity/internal/clarity"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func serviceResource() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A Clarity service.",

		CreateContext: serviceCreate,
		ReadContext:   serviceRead,
		DeleteContext: serviceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"provider_slug": {
				Description:  "Provider slug.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"name": {
				Description:  "A name for the service.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"unique": {
				Description: "Enforce a unique name for the service.",
				Type:        schema.TypeBool,
				Optional:    true,
				ForceNew:    true,
				Default:     false,
				DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
					return true
				},
				DiffSuppressOnRefresh: true,
			},
			"slug": {
				Description: "A slug for this service.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func serviceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)

	providerSlug := d.Get("provider_slug").(string)
	name := d.Get("name").(string)
	unique := d.Get("unique").(bool)

	// TODO better api support
	if unique {
		resp, err := client.ListServices()
		if err != nil {
			return diag.Errorf("loading services to confirm uniqueness: %v", err)
		}
		for _, s := range resp.Services {
			if s.Name == name {
				return diag.Errorf("The service name '%s' is not unique.", s.Name)
			}
		}
	}

	service, err := client.CreateService(clarity.ServiceCreateRequest{
		Name:               name,
		Resources:          make([]clarity.CreateResourceRequest, 0),
		RepositoryProvider: providerSlug,
		ServiceType:        "function",
	})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(service.Slug)
	d.Set("slug", service.Slug)

	tflog.Trace(ctx, "created a service")

	return nil
}

func serviceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)
	slug := d.Id()

	rsp, err := client.ListServices()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, service := range rsp.Services {
		if service.Slug == slug {
			d.Set("provider_slug", service.Provider.Slug)
			d.Set("name", service.Name)
			d.Set("slug", service.Slug)
			// TODO fix
			//			d.Set("unique", false)
			return nil
		}
	}

	return diag.Errorf("Not found")
}

func serviceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)
	slug := d.Id()
	err := client.DeleteService(slug)
	return diag.FromErr(err)
}
