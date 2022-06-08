package clarity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Resource struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Provider string `json:"provider"`
}

type InternalResource struct {
	Resource
	// For read
	Data       Configuration      `json:"data"`
	Deployment DeploymentStrategy `json:"deployment"`
}

func (x InternalResource) ManualUserInterfaceTrigger() bool {
	trigger := x.Deployment.Trigger[0]
	return trigger.Type == "event" && trigger.Event == "manual"
}

func (x InternalResource) updateStrategyRequest(single DeploymentRule) UpdateDeploymentStrategy {
	strategy := x.Deployment
	strategy.Trigger = []DeploymentRule{
		single,
	}

	return UpdateDeploymentStrategy{
		Strategy: strategy,
	}
}

func (x InternalResource) EnableUserInterfaceTrigger() UpdateDeploymentStrategy {
	return x.updateStrategyRequest(DeploymentRule{
		Type:  "event",
		Event: "manual",
	})
}

func (x InternalResource) DisableUserInterfaceTrigger() UpdateDeploymentStrategy {
	return x.updateStrategyRequest(DeploymentRule{
		Type:  "always",
		Event: "",
	})
}

type UpdateDeploymentStrategy struct {
	Strategy DeploymentStrategy `json:"strategy"`
}

type DeploymentStrategy struct {
	Trigger    []DeploymentRule `json:"trigger"`
	Health     []interface{}    `json:"health"`
	Evaluation []interface{}    `json:"evaluation"`
	Stages     []interface{}    `json:"stages"`
}

type DeploymentRule struct {
	Type  string `json:"type"`
	Event string `json:"name,omitempty"`
}

type CreateResourceRequest struct {
	Name          string        `json:"name"`
	Provider      string        `json:"provider"`     // slug
	RequestType   string        `json:"request_type"` // "import"
	Configuration Configuration `json:"configuration"`
}

type Configuration struct {
	Type string `json:"type"` // "lambda"
	// FIX bug here
	LambdaConfiguration LambdaConfiguration `json:"configuration"`
}

func (x Configuration) resourceName() string {
	// FIX aws
	return fmt.Sprintf("%s:%s", x.LambdaConfiguration.Name, x.LambdaConfiguration.Alias)
}

type LambdaConfiguration struct {
	Name  string `json:"name"`
	Alias string `json:"alias,omitempty"`
}

func (config *Client) ReadResource(serviceSlug string, resourceSlug string) (*InternalResource, error) {
	path := fmt.Sprintf("service/%s/resource/%s", serviceSlug, resourceSlug)
	statusCode, output, err := config.do(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	var res InternalResource
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return &res, nil
}

func (config *Client) CreateResource(serviceSlug string, rawreq CreateResourceRequest) (*InternalResource, error) {
	body, err := json.Marshal(rawreq)
	if err != nil {
		return nil, fmt.Errorf("Internal error creating request")
	}

	path := fmt.Sprintf("service/%s/resource", serviceSlug)
	statusCode, output, err := config.do(http.MethodPost, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if statusCode == http.StatusBadRequest {
		var res Error
		err = json.Unmarshal(output, &res)
		if err != nil {
			return nil, fmt.Errorf("Unhandled bad request decode: %w", err)
		}

		if res.Code == "service-candidate-not-found" {
			return nil, fmt.Errorf("Could not find the underlying resource '%s'", rawreq.Configuration.resourceName())
		}

		return nil, fmt.Errorf("Unhandled http-error-code: %v", res)
	}

	if statusCode == http.StatusNotFound {
		return nil, fmt.Errorf("Service slug not found.")
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	var res InternalResource
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return &res, nil
}

func (config *Client) DeleteResource(serviceSlug string, resourceSlug string) error {
	path := fmt.Sprintf("service/%s/resource/%s", serviceSlug, resourceSlug)
	statusCode, output, err := config.do(http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	if statusCode == http.StatusBadRequest {
		var res Error
		err = json.Unmarshal(output, &res)
		if err != nil {
			return fmt.Errorf("Unhandled bad request decode: %w", err)
		}

		if res.Code == "deployment-in-progress" {
			return fmt.Errorf("Unable to delete resource while there is an active deployment in progress")
		}

		return fmt.Errorf("Unhandled http-error-code: %v", res)
	}

	if statusCode == http.StatusNotFound {
		return fmt.Errorf("Service/Resource not found.")
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("unhandled http status code [%v]", statusCode)
	}
	return nil
}

func (config *Client) UpdateResourceDeploymentStrategy(serviceSlug string, resourceSlug string, rawreq UpdateDeploymentStrategy) error {
	body, err := json.Marshal(rawreq)
	if err != nil {
		return fmt.Errorf("Internal error creating request")
	}

	path := fmt.Sprintf("service/%s/resource/%s/strategy", serviceSlug, resourceSlug)
	statusCode, _, err := config.do(http.MethodPost, path, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if statusCode == http.StatusNotFound {
		return fmt.Errorf("Not found.")
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	return nil
}
