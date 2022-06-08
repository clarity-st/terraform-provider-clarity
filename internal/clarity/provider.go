package clarity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Provider struct {
	Name         string       `json:"name"`
	Slug         string       `json:"slug"`
	Info         ProviderInfo `json:"info"`
	Capabilities []string     `json:"capabilities"`
}

type TypeSwitch struct {
	Type string `json:"type"`
}

var AWSProviderType = TypeSwitch{"aws"}
var WebhookProviderType = TypeSwitch{"webhook"}

type ProviderInfo struct {
	TypeSwitch
	*AWS
	*Webhook
}

func (t *ProviderInfo) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &t.TypeSwitch); err != nil {
		return err
	}
	switch t.Type {
	case "aws":
		t.AWS = &AWS{}
		return json.Unmarshal(data, t.AWS)
	case "webhook":
		t.Webhook = &Webhook{}
		return json.Unmarshal(data, t.Webhook)
	default:
		return fmt.Errorf("unrecognized type value %q", t.Type)
	}

}

type AWS struct {
	AccountID           string  `json:"account_id"`
	AdditionalAccountID *string `json:"additional_account_id,omitempty"`
	Role                string  `json:"role"`
	Region              string  `json:"region"`
}

type Webhook struct {
	URL string `json:"url"`
}

func (config *Client) CreateProvider(name string, info ProviderInfo) (*Provider, error) {
	body, err := json.Marshal(struct {
		Name     string       `json:"name"`
		Provider ProviderInfo `json:"info"`
	}{
		Name:     name,
		Provider: info,
	})
	if err != nil {
		return nil, fmt.Errorf("Internal error creating request")
	}
	statusCode, output, err := config.do(http.MethodPost, "providers", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	var res Provider
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return &res, nil
}

func (config *Client) LoadProvider(slug string) (*Provider, error) {
	path := fmt.Sprintf("provider/%s", slug)
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

	var res Provider
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return &res, nil
}

func (config *Client) DeleteProvider(slug string) error {
	path := fmt.Sprintf("provider/%s", slug)
	statusCode, output, err := config.do(http.MethodDelete, path, nil)
	if statusCode == http.StatusBadRequest {
		var res Error
		err = json.Unmarshal(output, &res)
		if err != nil {
			return fmt.Errorf("Unhandled bad request decode: %w", err)
		}

		if res.Code == "provider-services-exist" {
			return fmt.Errorf("Unable to delete provider while there are services attached")
		}

		if res.Code == "provider-resources-exist" {
			return fmt.Errorf("Unable to delete provider while there are resources attached")
		}

		return fmt.Errorf("Unhandled http-error-code: %v", res)
	}

	if statusCode == http.StatusNotFound {
		return fmt.Errorf("Provider not found.")
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	return nil
}

func (config *Client) updateProviderName(slug string, name string) (*Provider, error) {
	body, err := json.Marshal(struct {
		Name string `json:"name"`
	}{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("Internal error creating request")
	}

	path := fmt.Sprintf("provider/%s", slug)
	statusCode, output, err := config.do(http.MethodPost, path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	var res Provider
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return &res, nil
}

func (config *Client) authenticateProvider(info ProviderInfo) error {
	body, err := json.Marshal(struct {
		Provider ProviderInfo `json:"info"`
	}{
		Provider: info,
	})
	if err != nil {
		return fmt.Errorf("Internal error creating request")
	}
	statusCode, _, err := config.do(http.MethodPost, "providers/authenticate", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	if statusCode == http.StatusForbidden {
		return fmt.Errorf("provider authentication failed, check your configuration")
	}

	if statusCode == http.StatusBadRequest {
		return fmt.Errorf("provider authentication failed, bad request")
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("provider authentication failed, unhandled http status code [%v]", statusCode)
	}

	return nil
}

func (config *Client) loadProviders() ([]Provider, error) {
	statusCode, output, err := config.do(http.MethodGet, "providers", nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	type providersListResponse struct {
		Providers []Provider `json:"providers"`
	}

	var res providersListResponse
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return res.Providers, nil
}
