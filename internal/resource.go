package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/clarity-st/terraform-provider-clarity/internal/clarity"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func renderID(service, resource string) string {
	return fmt.Sprintf("%s#%s", service, resource)
}

func parseID(input string) (string, string) {
	out := strings.Split(input, "#")
	service := out[0]
	resource := out[1]
	return service, resource
}

func resourceResource() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "A Clarity resource.",

		CreateContext: resourceCreate,
		ReadContext:   resourceRead,
		UpdateContext: resourceUpdate,
		DeleteContext: resourceDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"provider_slug": {
				Description: "Provider slug.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"service_slug": {
				Description: "Service slug.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"name": {
				Description: "A name for the resource.",
				Type:        schema.TypeString,
				ForceNew:    true,
				Required:    true,
			},
			"lambda": {
				Description: "AWS Lambda import configuration.",
				Type:        schema.TypeSet,
				MaxItems:    1,
				MinItems:    1,
				ForceNew:    true,
				Required:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"function_name": {
							Description:  "AWS Lambda function name",
							Type:         schema.TypeString,
							ForceNew:     true,
							Required:     true,
							ValidateFunc: validation.StringLenBetween(0, 255),
						},
						"alias": {
							Description: "AWS Lambda alias",
							Type:        schema.TypeString,
							ForceNew:    true,
							Optional:    true,
							Default:     "clarity",
						},
					},
				},
			},
			"deployment": {
				Description: "Deployment configuration.",
				Type:        schema.TypeSet,
				MaxItems:    1,
				Optional:    true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"trigger": {
							Description: "Deployment trigger configuration",
							Type:        schema.TypeSet,
							MaxItems:    1,
							Required:    true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"manual_user_interface": {
										Description: "Enable manual deployment lock, controlled via the user interface",
										Type:        schema.TypeBool,
										Optional:    true,
										Default:     false,
										DiffSuppressFunc: func(k, oldValue, newValue string, d *schema.ResourceData) bool {
											return oldValue == newValue
										},
									},
								},
							},
						},
					},
				},
			},

			"slug": {
				Description: "A slug for this resource.",
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func resourceCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*clarity.Client)

	providerSlug := d.Get("provider_slug").(string)
	serviceSlug := d.Get("service_slug").(string)
	name := d.Get("name").(string)

	lambdaSchema := d.Get("lambda").(*schema.Set)
	// Validation should enforce
	lambdaSchema0 := lambdaSchema.List()[0]
	lambda := lambdaSchema0.(map[string]interface{})
	functionName := lambda["function_name"].(string)
	alias := lambda["alias"].(string)

	// Validate
	service, err := api.LoadService(serviceSlug)
	if err != nil {
		return diag.Errorf("loading service '%s' for validation: %v", serviceSlug, err)
	}
	if service == nil {
		return diag.Errorf("Unable to find service with slug '%s'", serviceSlug)
	}
	for _, r := range service.Resources {
		if r.Name == name {
			return diag.Errorf("Conflict. Resource with the name '%s' already exists on the specified service", name)
		}
	}

	internal, err := api.CreateResource(serviceSlug, clarity.CreateResourceRequest{
		Name:        name,
		Provider:    providerSlug,
		RequestType: "import",
		Configuration: clarity.Configuration{
			Type: "aws",
			LambdaConfiguration: clarity.LambdaConfiguration{
				Name:  functionName,
				Alias: alias,
			},
		},
	})
	if err != nil {
		tflog.Error(ctx, "error", map[string]interface{}{
			"message": fmt.Sprintf("%v", err),
		})
		return diag.FromErr(err)
	}

	d.SetId(renderID(serviceSlug, internal.Slug))

	if v, ok := d.GetOk("deployment"); ok && len(v.(*schema.Set).List()) > 0 {
		deploymentSchema := v.(*schema.Set).List()[0]
		deployment := deploymentSchema.(map[string]interface{})
		triggerSchema := deployment["trigger"].(*schema.Set).List()[0]
		trigger := triggerSchema.(map[string]interface{})
		userInterfaceTrigger := trigger["manual_user_interface"].(bool)
		if userInterfaceTrigger {
			if err := api.UpdateResourceDeploymentStrategy(serviceSlug, internal.Slug, internal.EnableUserInterfaceTrigger()); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceRead(ctx, d, meta)
}

// Limited functionality constrianed to deployment triggers.
func resourceUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*clarity.Client)
	serviceSlug, resourceSlug := parseID(d.Id())

	if v, ok := d.GetOk("deployment"); ok && len(v.(*schema.Set).List()) > 0 {
		deploymentSchema := v.(*schema.Set).List()[0]
		deployment := deploymentSchema.(map[string]interface{})
		triggerSchema := deployment["trigger"].(*schema.Set).List()[0]
		trigger := triggerSchema.(map[string]interface{})
		userInterfaceTrigger := trigger["manual_user_interface"].(bool)

		internal, err := api.ReadResource(serviceSlug, resourceSlug)
		if err != nil {
			return diag.FromErr(err)
		}
		if internal.ManualUserInterfaceTrigger() != userInterfaceTrigger {
			tflog.Trace(ctx, fmt.Sprintf("updating user interface trigger: %v", userInterfaceTrigger))

			var strategy clarity.UpdateDeploymentStrategy
			if userInterfaceTrigger {
				strategy = internal.EnableUserInterfaceTrigger()
			} else {
				strategy = internal.DisableUserInterfaceTrigger()
			}

			if err := api.UpdateResourceDeploymentStrategy(serviceSlug, resourceSlug, strategy); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	return resourceRead(ctx, d, meta)
}

func resourceDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*clarity.Client)
	serviceSlug, resourceSlug := parseID(d.Id())

	err := api.DeleteResource(serviceSlug, resourceSlug)
	if err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func lambdaHash(in interface{}) int {
	var buf bytes.Buffer
	m := in.(map[string]interface{})
	for k, v := range m {
		buf.WriteString(fmt.Sprintf("%s#%s", k, v.(string)))
	}

	return HashString(buf.String())
}

func deploymentHash(in interface{}) int {
	var buf bytes.Buffer
	m := in.(map[string]interface{})

	triggerInterface := m["trigger"]
	triggerSchema := triggerInterface.(*schema.Set).List()[0]
	trigger := triggerSchema.(map[string]interface{})
	if userInterface, ok := trigger["manual_user_interface"].(bool); ok {
		buf.WriteString(fmt.Sprintf("manual_user_interface#%v", userInterface))
	}

	return HashString(buf.String())
}

func triggerHash(in interface{}) int {
	var buf bytes.Buffer
	m := in.(map[string]interface{})
	if userInterface, ok := m["manual_user_interface"]; ok {
		buf.WriteString(fmt.Sprintf("manual_user_interface#%v", userInterface))
	}

	return HashString(buf.String())
}

func resourceRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	api := meta.(*clarity.Client)
	serviceSlug, resourceSlug := parseID(d.Id())

	internal, err := api.ReadResource(serviceSlug, resourceSlug)
	if err != nil {
		if errors.Is(err, clarity.ErrNotFound) {
			d.SetId("")
			return nil
		}

		return diag.FromErr(err)
	}

	d.Set("provider_slug", internal.Provider)
	d.Set("service_slug", serviceSlug)
	d.Set("name", internal.Name)

	// TODO Support more resource types
	switch internal.Data.Type {
	case "lambda":
		lambdaConf := internal.Data.LambdaConfiguration
		d.Set("lambda", schema.NewSet(lambdaHash, []interface{}{
			map[string]interface{}{
				"function_name": lambdaConf.Name,
				"alias":         lambdaConf.Alias,
			},
		}))
	}

	// TODO Support full deployment configuration
	if len(internal.Deployment.Trigger) == 1 {
		d.Set("deployment", schema.NewSet(deploymentHash, []interface{}{
			map[string]interface{}{
				"trigger": schema.NewSet(triggerHash, []interface{}{
					map[string]interface{}{
						"manual_user_interface": internal.ManualUserInterfaceTrigger(),
					},
				}),
			},
		}))
	}

	d.Set("slug", internal.Slug)

	return nil
}
