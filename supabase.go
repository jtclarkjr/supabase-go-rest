package supabase

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// Client represents the Supabase client
type Client struct {
	BaseUrl string
	ApiKey  string
	Token   string
}

const restApiPath = "/rest/v1"

// NewClient creates a new Supabase client
func NewClient(baseUrl, apiKey, token string) *Client {
	return &Client{
		BaseUrl: baseUrl,
		ApiKey:  apiKey,
		Token:   token,
	}
}

// Get performs a GET request to the Supabase REST API. Requires table name and query param.
func (c *Client) Get(endpoint string, queryParams ...map[string]string) ([]byte, error) {
	params := map[string]string{}
	if len(queryParams) > 0 {
		params = queryParams[0]
	}
	return c.doRequest("GET", endpoint, params, nil)
}

// Post performs a POST request to the Supabase REST API. Requires table name, and request data.
func (c *Client) Post(endpoint string, data []byte) ([]byte, error) {
	return c.doRequest("POST", endpoint, nil, bytes.NewBuffer(data))
}

// Put performs a PUT request to the Supabase REST API. Requires table name, primary key, primary key value, and request data.
func (c *Client) Put(endpoint string, primaryKeyName string, primaryKeyValue string, data []byte) ([]byte, error) {
	query := map[string]string{
		primaryKeyName: primaryKeyValue,
	}
	return c.doRequest("PUT", endpoint, query, bytes.NewBuffer(data))
}

// Delete performs a DELETE request to the Supabase REST API. Requires table name, primary key, and primary key value.
func (c *Client) Delete(endpoint string, primaryKeyName string, primaryKeyValue string) ([]byte, error) {
	query := map[string]string{
		primaryKeyName: primaryKeyValue,
	}
	return c.doRequest("DELETE", endpoint, query, nil)
}

// formatQueryParams formats query parameters for Supabase compatibility
func formatQueryParams(params map[string]string) map[string]string {
	formattedParams := make(map[string]string)
	for key, value := range params {
		formattedParams[key] = fmt.Sprintf("eq.%s", url.QueryEscape(value))
	}
	return formattedParams
}

// doRequest performs the actual HTTP request. Requires API key, and Token for headers
func (c *Client) doRequest(method, endpoint string, queryParams map[string]string, body io.Reader) ([]byte, error) {
	urlStr := fmt.Sprintf("%s%s/%s", c.BaseUrl, restApiPath, endpoint)
	if len(queryParams) > 0 {
		urlObj, err := url.Parse(urlStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse URL: %v", err)
		}
		q := urlObj.Query()
		for key, value := range formatQueryParams(queryParams) {
			q.Add(key, value)
		}
		urlObj.RawQuery = q.Encode()
		urlStr = urlObj.String()
	}

	req, err := http.NewRequest(method, urlStr, body)
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
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}
