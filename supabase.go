package supabase

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

/*
 **********************
 *   TYPE DEFINITON   *
 **********************
 */

// Client represents the Supabase client
type Client struct {
	BaseUrl string
	ApiKey  string
	Token   string
}

// AuthTokenResponse represents the response from the /token endpoint
type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

// TokenRequestPayload represents the payload for /token requests
type TokenRequestPayload struct {
	Email        string `json:"email,omitempty"`
	Password     string `json:"password,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
	GrantType    string `json:"grant_type,omitempty"`
}

// MagicLinkPayload represents the payload for sending magic links
type MagicLinkPayload struct {
	Email string `json:"email"`
}

// VerifyOTPPayload represents the payload for verifying OTP
type VerifyOTPPayload struct {
	Email string `json:"email"`
	Token string `json:"token"`
	Type  string `json:"type"`
}

/*
 **************************
 *   VARIABLE DEFINITON   *
 **************************
 */

// Defined REST API paths from Supabase
const (
	restApiPath      = "/rest/v1"
	authApiPath      = "/auth/v1"
	tokenApiPath     = authApiPath + "/token"
	signupApiPath    = authApiPath + "/signup"
	magicLinkApiPath = authApiPath + "/magiclink"
	recoverApiPath   = authApiPath + "/recover"
	verifyApiPath    = authApiPath + "/verify"
	userApiPath      = authApiPath + "/user"
	logoutApiPath    = authApiPath + "/logout"
	inviteApiPath    = authApiPath + "/invite"
	resetApiPath     = authApiPath + "/reset"
)

// Custom error types
var (
	ErrInvalidResponse = errors.New("invalid response from server")
	ErrRequestFailed   = errors.New("request failed")
)

/*
 *******************
 *   INITIALIZER   *
 *******************
 */

// NewClient creates a new Supabase client
func NewClient(baseUrl, apiKey, token string) *Client {
	return &Client{
		BaseUrl: baseUrl,
		ApiKey:  apiKey,
		Token:   token,
	}
}

/*
 ********************
 *   AUTH METHODS   *
 ********************
 */

// SignUp creates a new user
func (c *Client) SignUp(email, password string) ([]byte, error) {
	payload := map[string]string{"email": email, "password": password}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s?grant_type=signup", signupApiPath)

	response, err := c.doRequest("POST", path, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("SignUp error: %v", err)
		return nil, err
	}
	return response, nil
}

// SignIn authenticates a user and retrieves a token
func (c *Client) SignIn(email, password string) (*AuthTokenResponse, error) {
	payload := TokenRequestPayload{
		Email:    email,
		Password: password,
	}

	path := fmt.Sprintf("%s?grant_type=password", tokenApiPath)

	authResponse, err := c.authRequest(path, payload)
	if err != nil {
		log.Printf("SignIn error: %v", err)
		return nil, err
	}
	return authResponse, nil
}

// RefreshToken refreshes the access token
func (c *Client) RefreshToken(refreshToken string) (*AuthTokenResponse, error) {
	payload := TokenRequestPayload{
		RefreshToken: refreshToken,
	}

	path := fmt.Sprintf("%s?grant_type=refresh_token", tokenApiPath)

	authResponse, err := c.authRequest(path, payload)
	if err != nil {
		log.Printf("RefreshToken error: %v", err)
		return nil, err
	}
	return authResponse, nil
}

// SendMagicLink sends a magic link to the user's email
func (c *Client) SendMagicLink(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest("POST", magicLinkApiPath, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("SendMagicLink error: %v", err)
		return nil, err
	}
	return response, nil
}

// SendPasswordRecovery sends a password recovery email
func (c *Client) SendPasswordRecovery(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest("POST", recoverApiPath, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("SendPasswordRecovery error: %v", err)
		return nil, err
	}
	return response, nil
}

// VerifyOTP verifies a one-time password (OTP)
func (c *Client) VerifyOTP(email, token, otpType string) ([]byte, error) {
	payload := VerifyOTPPayload{
		Email: email,
		Token: token,
		Type:  otpType,
	}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.doRequest("POST", verifyApiPath, nil, bytes.NewBuffer(jsonData))
}

// GetUser retrieves the authenticated user's information
func (c *Client) GetUser() ([]byte, error) {
	return c.doRequest("GET", userApiPath, nil, nil)
}

// UpdateUser updates the authenticated user's information
func (c *Client) UpdateUser(payload map[string]string) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.doRequest("PUT", userApiPath, nil, bytes.NewBuffer(jsonData))
}

// SignOut logs out the user
func (c *Client) SignOut() ([]byte, error) {
	return c.doRequest("POST", logoutApiPath, nil, nil)
}

// InviteUser sends an invite email to a new user (admin only)
func (c *Client) InviteUser(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return c.doRequest("POST", inviteApiPath, nil, bytes.NewBuffer(jsonData))
}

// ResetPassword resets the user's password using a token
func (c *Client) ResetPassword(token, newPassword string) ([]byte, error) {
	payload := map[string]string{
		"token":    token,
		"password": newPassword,
	}

	path := fmt.Sprintf("%s?grant_type=reset_password", resetApiPath)

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	response, err := c.doRequest("POST", path, nil, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("ResetPassword error: %v", err)
		return nil, err
	}
	return response, nil
}

/*
 ********************
 *   HTTP METHODS   *
 ********************
 */

// Get performs a GET request to the Supabase REST API. Requires table name, and query parameters.
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

/*
 ********************
 *   REQ METHODS   *
 ********************
 */

// formatQueryParams formats query parameters for Supabase compatibility
func formatQueryParams(params map[string]string) map[string]string {
	formattedParams := make(map[string]string)
	for key, value := range params {
		formattedParams[key] = fmt.Sprintf("eq.%s", url.QueryEscape(value))
	}
	return formattedParams
}

// authRequest handles authentication-related requests
func (c *Client) authRequest(endpoint string, payload TokenRequestPayload) (*AuthTokenResponse, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		log.Printf("authRequest: failed to marshal payload: %v", err)
		return nil, errors.New("failed to marshal payload")
	}

	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, endpoint)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("authRequest: failed to create request: %v", err)
		return nil, errors.New("failed to create request")
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("authRequest: failed to perform request: %v", err)
		return nil, errors.New("failed to perform request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("authRequest: request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, errors.New("request failed")
	}

	var authResponse AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		log.Printf("authRequest: failed to decode response: %v", err)
		return nil, errors.New("failed to decode response")
	}

	return &authResponse, nil
}

// doRequest performs the actual HTTP request. Requires API key, and Token for headers
func (c *Client) doRequest(method, endpoint string, queryParams map[string]string, body io.Reader) ([]byte, error) {
	// Normalize endpoint to avoid double slashes when endpoint begins with '/'
	cleanEndpoint := strings.TrimPrefix(endpoint, "/")
	urlStr := fmt.Sprintf("%s%s/%s", c.BaseUrl, restApiPath, cleanEndpoint)
	if len(queryParams) > 0 {
		urlObj, err := url.Parse(urlStr)
		if err != nil {
			log.Printf("Error: doRequest failed to parse URL - %v", err)
			return nil, nil
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
		log.Printf("Error: doRequest failed to create request - %v", err)
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)

	if c.Token != "" {
		if !strings.HasPrefix(c.Token, "Bearer ") {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
		} else {
			req.Header.Set("Authorization", c.Token)
		}
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error: doRequest failed to perform request - %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Error: doRequest %v - %s", ErrRequestFailed, string(body))
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error: doRequest failed to read response body - %v", err)
	}

	return responseBody, nil
}
