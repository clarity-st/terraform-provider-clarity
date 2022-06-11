package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/clarity-st/terraform-provider-clarity/internal/clarity"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// var regionRegexp = regexp.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d$`)

func providerResource() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A Clarity provider.",

		CreateContext: providerCreate,
		ReadContext:   providerRead,
		DeleteContext: providerDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Description:  "A name for the provider.",
				Type:         schema.TypeString,
				ForceNew:     true,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(0, 128),
			},
			"aws": {
				Description:   "AWS Provider configuration.",
				Type:          schema.TypeSet,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"webhook"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"account_id": {
							Description:  "AWS Account ID",
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(12, 12),
						},
						"additional_account_id": {
							Description:  "Additional AWS Account ID",
							Type:         schema.TypeString,
							ForceNew:     true,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(12, 12),
						},
						"role": {
							Description:  "IAM Role name",
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"region": {
							Description:  "AWS Region",
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 20),
						},
					},
				},
			},
			"webhook": {
				Description:   "Webhook configuration.",
				Type:          schema.TypeSet,
				MaxItems:      1,
				MinItems:      1,
				ForceNew:      true,
				Optional:      true,
				ConflictsWith: []string{"aws"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"url": {
							Description:  "URL",
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 250),
						},
					},
				},
			},
			"slug": {
				Description: "A slug for this provider.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func providerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)

	info := clarity.ProviderInfo{}
	name := d.Get("name").(string)

	var typeSet int
	if v, ok := d.GetOk("aws"); ok && len(v.(*schema.Set).List()) > 0 {
		aws := v.(*schema.Set).List()[0].(map[string]interface{})
		account_id := aws["account_id"].(string)
		var additional_account_id *string
		if v, ok := aws["additional_account_id"]; ok {
			s := v.(string)
			additional_account_id = &s
		}
		role := aws["role"].(string)
		region := aws["region"].(string)

		info.TypeSwitch = clarity.AWSProviderType
		info.AWS = &clarity.AWS{
			AccountID:           account_id,
			AdditionalAccountID: additional_account_id,
			Role:                role,
			Region:              region,
		}
		typeSet++
	}

	if v, ok := d.GetOk("webhook"); ok && len(v.(*schema.Set).List()) > 0 {
		webhook := v.(*schema.Set).List()[0].(map[string]interface{})
		url := webhook["url"].(string)

		info.TypeSwitch = clarity.WebhookProviderType
		info.Webhook = &clarity.Webhook{
			URL: url,
		}
		typeSet++
	}

	if typeSet != 1 {
		return diag.Errorf("Must specific exactly one of 'aws' or 'webhook'")
	}

	providers, err := client.LoadProviders()
	if err != nil {
		return diag.Errorf("loading provider to confirm uniqueness: %v", err)
	}

	for _, p := range providers {
		if p.Name == name {
			return diag.Errorf("Conflict. A provider with the name '%s' already exists.", name)
		}
	}

	provider, err := client.CreateProvider(name, info)
	if err != nil {
		return diag.Errorf("creating provider: %v", err)
	}

	d.SetId(provider.Slug)

	tflog.Trace(ctx, "created a new provider")

	return providerRead(ctx, d, meta)
}

func mapHash(in interface{}) int {
	var buf bytes.Buffer
	m := in.(map[string]interface{})
	for k, i := range m {
		switch v := i.(type) {
		case string:
			buf.WriteString(fmt.Sprintf("%s#%s", k, v))
		case *string:
			buf.WriteString(fmt.Sprintf("%s#%s", k, *v))
		default:
			buf.WriteString(fmt.Sprintf("%s#%v", k, v))
		}
	}

	return HashString(buf.String())
}

func providerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)
	slug := d.Id()

	rsp, err := client.LoadProvider(slug)
	if err != nil {
		if errors.Is(err, clarity.ErrNotFound) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	d.Set("name", rsp.Name)
	if rsp.Info.AWS != nil {
		s := map[string]interface{}{
			"account_id": rsp.Info.AWS.AccountID,
			"role":       rsp.Info.AWS.Role,
			"region":     rsp.Info.AWS.Region,
		}

		if rsp.Info.AWS.AdditionalAccountID != nil && len(*rsp.Info.AWS.AdditionalAccountID) > 0 {
			s["additional_account_id"] = rsp.Info.AWS.AdditionalAccountID
		}
		d.Set("aws", schema.NewSet(mapHash, []interface{}{s}))
	}
	if rsp.Info.Webhook != nil {
		d.Set("webhook", schema.NewSet(mapHash, []interface{}{
			map[string]interface{}{
				"url": rsp.Info.Webhook.URL,
			},
		}))
	}

	d.Set("slug", rsp.Slug)

	return nil
}

func providerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*clarity.Client)
	slug := d.Id()
	err := client.DeleteProvider(slug)
	return diag.FromErr(err)
}
