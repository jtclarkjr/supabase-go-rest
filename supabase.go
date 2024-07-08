package supabase

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

// TClient represents the Supabase client
type TClient struct {
	BaseUrl string
	ApiKey  string
	Token   string
}

const restApiPath = "/rest/v1"

// NewClient creates a new Supabase client
func NewClient(baseUrl, apiKey, token string) *TClient {
	return &TClient{
		BaseUrl: baseUrl,
		ApiKey:  apiKey,
		Token:   token,
	}
}

// Get performs a GET request to the Supabase REST API
func (c *TClient) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest("GET", endpoint, queryParams, nil)
}

// Post performs a POST request to the Supabase REST API
func (c *TClient) Post(endpoint string, data []byte) ([]byte, error) {
	return c.doRequest("POST", endpoint, nil, bytes.NewBuffer(data))
}

// Put performs a PUT request to the Supabase REST API
func (c *TClient) Put(endpoint string, data []byte) ([]byte, error) {
	return c.doRequest("PUT", endpoint, nil, bytes.NewBuffer(data))
}

// Delete performs a DELETE request to the Supabase REST API
func (c *TClient) Delete(endpoint string) ([]byte, error) {
	return c.doRequest("DELETE", endpoint, nil, nil)
}

// doRequest performs the actual HTTP request
func (c *TClient) doRequest(method, endpoint string, queryParams map[string]string, body io.Reader) ([]byte, error) {
	url := fmt.Sprintf("%s%s/%s", c.BaseUrl, restApiPath, endpoint)
	if queryParams != nil {
		q := url + "?"
		for key, value := range queryParams {
			q += fmt.Sprintf("%s=%s&", key, value)
		}
		url = q[:len(q)-1] // remove the trailing '&'
	}

	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Authorization", c.Token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}
