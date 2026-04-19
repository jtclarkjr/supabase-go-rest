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
| **********************
| *   TYPE DEFINITON   *
| **********************
| */

// Client represents the Supabase client
type Client struct {
	BaseUrl string
	ApiKey  string
	Token   string
}

// QueryBuilder represents a fluent query builder
type QueryBuilder struct {
	client      *Client
	table       string
	queryParams map[string]string
	method      string
	body        []byte
}

// AuthTokenResponse represents the response from the /token endpoint
type AuthTokenResponse struct {
	AccessToken          string `json:"access_token"`
	TokenType            string `json:"token_type"`
	ExpiresIn            int    `json:"expires_in"`
	RefreshToken         string `json:"refresh_token"`
	ProviderToken        string `json:"provider_token,omitempty"`
	ProviderRefreshToken string `json:"provider_refresh_token,omitempty"`
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
| **************************
| *   VARIABLE DEFINITON   *
| **************************
| */

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
	authorizeApiPath = authApiPath + "/authorize"
)

// Custom error types
var (
	ErrInvalidResponse = errors.New("invalid response from server")
	ErrRequestFailed   = errors.New("request failed")
)

/*
| *******************
| *   INITIALIZER   *
| *******************
| */

// NewClient creates a new Supabase client
func NewClient(baseUrl, apiKey, token string) *Client {
	return &Client{
		BaseUrl: baseUrl,
		ApiKey:  apiKey,
		Token:   token,
	}
}

// From initiates a query on a specific table
func (c *Client) From(table string) *QueryBuilder {
	return &QueryBuilder{
		client:      c,
		table:       table,
		queryParams: make(map[string]string),
		method:      "GET",
	}
}

/*
| ********************
| *   AUTH METHODS   *
| ********************
| */

// SignUp creates a new user
func (c *Client) SignUp(email, password string) ([]byte, error) {
	payload := map[string]string{"email": email, "password": password}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("%s%s?grant_type=signup", c.BaseUrl, signupApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("SignUp error: %v", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("SignUp: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SignUp error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// SignInAnonymously creates an anonymous user session.
// Anonymous sign-ins must be enabled in your Supabase project settings.
func (c *Client) SignInAnonymously() (*AuthTokenResponse, error) {
	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, signupApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBufferString("{}"))
	if err != nil {
		return nil, fmt.Errorf("SignInAnonymously: failed to create request: %w", err)
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("SignInAnonymously: request failed: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("SignInAnonymously: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("SignInAnonymously: request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, ErrRequestFailed
	}

	var authResponse AuthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResponse); err != nil {
		return nil, fmt.Errorf("SignInAnonymously: failed to decode response: %w", err)
	}
	return &authResponse, nil
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

	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, magicLinkApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("SendMagicLink error: %v", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("SendMagicLink: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SendMagicLink error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// SendPasswordRecovery sends a password recovery email
func (c *Client) SendPasswordRecovery(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, recoverApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("SendPasswordRecovery error: %v", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("SendPasswordRecovery: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SendPasswordRecovery error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
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

	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, verifyApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("VerifyOTP: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("VerifyOTP error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// GetUser retrieves the authenticated user's information
func (c *Client) GetUser() ([]byte, error) {
	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, userApiPath)
	req, err := http.NewRequest("GET", urlStr, nil)
	if err != nil {
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("GetUser: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("GetUser error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// UpdateUser updates the authenticated user's information
func (c *Client) UpdateUser(payload map[string]string) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, userApiPath)
	req, err := http.NewRequest("PUT", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		if !strings.HasPrefix(c.Token, "Bearer ") {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
		} else {
			req.Header.Set("Authorization", c.Token)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("UpdateUser: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("UpdateUser error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// SignOut logs out the user
func (c *Client) SignOut() ([]byte, error) {
	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, logoutApiPath)
	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("SignOut: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("SignOut error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// InviteUser sends an invite email to a new user (admin only)
func (c *Client) InviteUser(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("%s%s", c.BaseUrl, inviteApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")
	if c.Token != "" {
		if !strings.HasPrefix(c.Token, "Bearer ") {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
		} else {
			req.Header.Set("Authorization", c.Token)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("InviteUser: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("InviteUser error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

// ResetPassword resets the user's password using a token
func (c *Client) ResetPassword(token, newPassword string) ([]byte, error) {
	payload := map[string]string{
		"token":    token,
		"password": newPassword,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	urlStr := fmt.Sprintf("%s%s?grant_type=reset_password", c.BaseUrl, resetApiPath)
	req, err := http.NewRequest("POST", urlStr, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("apikey", c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("ResetPassword error: %v", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("ResetPassword: error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("ResetPassword error: status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	return io.ReadAll(resp.Body)
}

/*
| ********************
| *   REQ METHODS   *
| ********************
| */

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
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

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

/*
| *************************
| *   QUERY BUILDER API   *
| *************************
| */

// Select specifies columns to return (use "*" for all)
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	if len(columns) > 0 && columns[0] != "" {
		qb.queryParams["select"] = columns[0]
	} else {
		qb.queryParams["select"] = "*"
	}
	return qb
}

// Insert adds data for insertion
func (qb *QueryBuilder) Insert(data interface{}) *QueryBuilder {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Insert: failed to marshal data: %v", err)
		return qb
	}
	qb.method = "POST"
	qb.body = jsonData
	return qb
}

// Update adds data for update
func (qb *QueryBuilder) Update(data interface{}) *QueryBuilder {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Printf("Update: failed to marshal data: %v", err)
		return qb
	}
	qb.method = "PATCH"
	qb.body = jsonData
	return qb
}

// Delete marks the query as a delete operation
func (qb *QueryBuilder) Delete() *QueryBuilder {
	qb.method = "DELETE"
	return qb
}

// Eq adds an equality filter
func (qb *QueryBuilder) Eq(column, value string) *QueryBuilder {
	qb.queryParams[column] = fmt.Sprintf("eq.%s", url.QueryEscape(value))
	return qb
}

// Order adds ordering to the query
func (qb *QueryBuilder) Order(column string, opts map[string]bool) *QueryBuilder {
	orderValue := column
	if opts != nil {
		if ascending, exists := opts["ascending"]; exists && !ascending {
			orderValue += ".desc"
		} else {
			orderValue += ".asc"
		}
	} else {
		orderValue += ".asc"
	}
	qb.queryParams["order"] = orderValue
	return qb
}

// Limit adds a limit to the query
func (qb *QueryBuilder) Limit(count int) *QueryBuilder {
	qb.queryParams["limit"] = fmt.Sprintf("%d", count)
	return qb
}

// Single expects a single row result
func (qb *QueryBuilder) Single() *QueryBuilder {
	qb.queryParams["limit"] = "1"
	return qb
}

// Execute runs the query and returns the result
func (qb *QueryBuilder) Execute() ([]byte, error) {
	var body io.Reader
	if qb.body != nil {
		body = bytes.NewBuffer(qb.body)
	}

	// Build URL
	cleanTable := strings.TrimPrefix(qb.table, "/")
	urlStr := fmt.Sprintf("%s%s/%s", qb.client.BaseUrl, restApiPath, cleanTable)

	if len(qb.queryParams) > 0 {
		urlObj, err := url.Parse(urlStr)
		if err != nil {
			log.Printf("Execute: failed to parse URL - %v", err)
			return nil, err
		}
		q := urlObj.Query()
		for key, value := range qb.queryParams {
			// Don't double-format if already formatted
			if key != "select" && key != "order" && key != "limit" && !strings.HasPrefix(value, "eq.") {
				value = fmt.Sprintf("eq.%s", url.QueryEscape(value))
			}
			q.Add(key, value)
		}
		urlObj.RawQuery = q.Encode()
		urlStr = urlObj.String()
	}

	req, err := http.NewRequest(qb.method, urlStr, body)
	if err != nil {
		log.Printf("Execute: failed to create request - %v", err)
		return nil, err
	}

	req.Header.Set("apikey", qb.client.ApiKey)

	if qb.client.Token != "" {
		if !strings.HasPrefix(qb.client.Token, "Bearer ") {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", qb.client.Token))
		} else {
			req.Header.Set("Authorization", qb.client.Token)
		}
	}

	req.Header.Set("Content-Type", "application/json")

	// Add Prefer header for insert/update to return data
	if qb.method == "POST" || qb.method == "PATCH" {
		if _, hasSelect := qb.queryParams["select"]; hasSelect || qb.method == "POST" {
			req.Header.Set("Prefer", "return=representation")
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Execute: failed to perform request - %v", err)
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("Execute: request failed with status %d: %s", resp.StatusCode, string(body))
		return nil, ErrRequestFailed
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Execute: failed to read response body - %v", err)
		return nil, err
	}

	return responseBody, nil
}
