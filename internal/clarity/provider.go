package clarity

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Provider struct {
	Name         string      `json:"name"`
	Slug         string      `json:"slug"`
	Info         interface{} `json:"info"`
	Capabilities []string    `json:"capabilities"`
}

func (config *Client) loadProvider(slug string) (*Provider, error) {
	providers, err := config.loadProviders()
	if err != nil {
		return nil, err
	}

	for _, provider := range providers {
		if provider.Slug == slug {
			return &provider, nil
		}
	}

	return nil, fmt.Errorf("unimplemented")
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
