package supabase

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Client represents the Supabase client
type Client struct {
	BaseUrl string
	ApiKey  string
	Token   string
}

type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// TokenRequestPayload represents the payload for /token requests
type TokenRequestPayload struct {
	Email        string `json:"email,omitempty"`
	Phone        string `json:"phone,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	GrantType    string `json:"grant_type"`
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

// Patch performs a PATCH request to the Supabase REST API. Requires table name, query parameters, and request data.
func (c *Client) Patch(endpoint string, queryParams map[string]string, data []byte) ([]byte, error) {
	return c.doRequest("PATCH", endpoint, queryParams, bytes.NewBuffer(data))
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

// AuthToken performs a POST request to the /token endpoint for authentication
func (c *Client) AuthToken(payload TokenRequestPayload) (*AuthTokenResponse, error) {
	endpoint := "/token"

	// Prepare the request body
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %v", err)
	}

	urlStr := fmt.Sprintf("%s%s%s", c.BaseUrl, restApiPath, endpoint)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Use the API key stored in the Client struct
	req.Header.Set("apikey", c.ApiKey)
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

	// Parse the response
	var authResponse AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &authResponse, nil
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

	// Set Authorization header only if token is provided
	if c.Token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform request: %v", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("error: %s", string(body))
	}

	return io.ReadAll(resp.Body)
}
