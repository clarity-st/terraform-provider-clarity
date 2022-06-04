package internal

import (
	"context"
	"fmt"
	"net/http"

	"github.com/clarity-st/terraform-provider-clarity/internal/clarity"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		p := &schema.Provider{
			Schema: map[string]*schema.Schema{
				"clarity_api_token": &schema.Schema{
					Type:        schema.TypeString,
					Optional:    true,
					Sensitive:   true,
					DefaultFunc: schema.EnvDefaultFunc("CLARITY_API_TOKEN", nil),
				},
				"service_endpoint": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Address of the Clarity service endpoint to use.",
					Default:     "https://api.clarity.st",
				},
			},
			//			DataSourcesMap: map[string]*schema.Resource{
			//				"clarity_provider": dataSourceProvider(),
			//			},
			ResourcesMap: map[string]*schema.Resource{
				"clarity_service":  serviceResource(),
				"clarity_resource": resourceResource(),
				"clarity_provider": providerResource(),
			},
		}

		p.ConfigureContextFunc = configure(version, p)
		return p
	}
}

func configure(version string, p *schema.Provider) func(context.Context, *schema.ResourceData) (interface{}, diag.Diagnostics) {
	return func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		host := d.Get("service_endpoint").(string)
		accessToken := d.Get("clarity_api_token").(string)

		var diags diag.Diagnostics

		if accessToken == "" {
			return nil, diag.FromErr(fmt.Errorf("'clarity_api_token' must be specified and not empty."))
		}

		return &clarity.Client{
			Host:      host,
			Token:     accessToken,
			UserAgent: p.UserAgent("terraform-provider-clarity", version),
			Client:    &http.Client{},
		}, diags
	}
}
