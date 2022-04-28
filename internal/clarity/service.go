package clarity

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceCreateRequest struct {
	Name               string                  `json:"name"`
	Resources          []CreateResourceRequest `json:"resources"`
	RepositoryProvider string                  `json:"repository_provider"`
	ServiceType        string                  `json:"type"`
}

type ServicesListResponse struct {
	Services []Service `json:"services"`
}

type Service struct {
	Name        string     `json:"name"`
	Slug        string     `json:"slug"`
	Resources   []Resource `json:"resources"`
	Provider    *Provider  `json:"repository_provider"`
	ServiceType string     `json:"type"`
}

func (config *Client) CreateService(rawreq ServiceCreateRequest) (*Service, error) {
	body, err := json.Marshal(rawreq)
	if err != nil {
		return nil, fmt.Errorf("Internal error creating request")
	}

	statusCode, output, err := config.do(http.MethodPost, "services", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	var res Service
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode resposne from server: %w", err)
	}

	return &res, nil
}

func (config *Client) DeleteService(serviceSlug string) error {
	statusCode, output, err := config.do(http.MethodDelete, fmt.Sprintf("service/%s", serviceSlug), nil)
	if err != nil {
		return err
	}

	if statusCode == http.StatusBadRequest {
		var res Error
		err = json.Unmarshal(output, &res)
		if err != nil {
			return fmt.Errorf("Unhandled bad request decode: %w", err)
		}

		if res.Code == "service-resources-exist" {
			return fmt.Errorf("The service has resources defined, you must delete all resources before you can delete the service.")
		}

		return fmt.Errorf("Unhandled http-error-code: %v", res)
	}

	if statusCode == http.StatusNotFound {
		return fmt.Errorf("Service slug not found.")
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("unhandled http status code [%v]", statusCode)
	}
	return nil
}

func (config *Client) ListServices() (*ServicesListResponse, error) {
	statusCode, output, err := config.do(http.MethodGet, "services", nil)
	if err != nil {
		return nil, err
	}

	if statusCode != http.StatusOK {
		return nil, fmt.Errorf("unhandled http status code [%v]", statusCode)
	}

	var res ServicesListResponse
	err = json.Unmarshal(output, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response from server: %w", err)
	}

	return &res, nil
}
