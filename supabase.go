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

// NewClient creates a new Supabase client
func NewClient(baseUrl, apiKey, token string) *Client {
	return &Client{
		BaseUrl: baseUrl,
		ApiKey:  apiKey,
		Token:   token,
	}
}

// SignUp creates a new user
func (c *Client) SignUp(email, password string) ([]byte, error) {
	payload := map[string]string{"email": email, "password": password}

	path := fmt.Sprintf("%s?grant_type=signup", signupApiPath)

	response, err := c.doRequest("POST", path, nil, payload)
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
	response, err := c.doRequest("POST", magicLinkApiPath, nil, payload)
	if err != nil {
		log.Printf("SendMagicLink error: %v", err)
		return nil, err
	}
	return response, nil
}

// SendPasswordRecovery sends a password recovery email
func (c *Client) SendPasswordRecovery(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	response, err := c.doRequest("POST", recoverApiPath, nil, payload)
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
	return c.doRequest("POST", verifyApiPath, nil, payload)
}

// GetUser retrieves the authenticated user's information
func (c *Client) GetUser() ([]byte, error) {
	return c.doRequest("GET", userApiPath, nil, nil)
}

// UpdateUser updates the authenticated user's information
func (c *Client) UpdateUser(payload map[string]string) ([]byte, error) {
	return c.doRequest("PUT", userApiPath, nil, payload)
}

// SignOut logs out the user
func (c *Client) SignOut() ([]byte, error) {
	return c.doRequest("POST", logoutApiPath, nil, nil)
}

// InviteUser sends an invite email to a new user (admin only)
func (c *Client) InviteUser(email string) ([]byte, error) {
	payload := MagicLinkPayload{Email: email}
	return c.doRequest("POST", inviteApiPath, nil, payload)
}

// ResetPassword resets the user's password using a token
func (c *Client) ResetPassword(token, newPassword string) ([]byte, error) {
	payload := map[string]string{
		"token":    token,
		"password": newPassword,
	}

	path := fmt.Sprintf("%s?grant_type=reset_password", resetApiPath)

	response, err := c.doRequest("POST", path, nil, payload)
	if err != nil {
		log.Printf("ResetPassword error: %v", err)
		return nil, err
	}
	return response, nil
}

// Get performs a GET request
func (c *Client) Get(endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest("GET", endpoint, queryParams, nil)
}

// Post performs a POST request
func (c *Client) Post(endpoint string, payload any) ([]byte, error) {
	return c.doRequest("POST", endpoint, nil, payload)
}

// Put performs a PUT request
func (c *Client) Put(endpoint string, payload any) ([]byte, error) {
	return c.doRequest("PUT", endpoint, nil, payload)
}

// Patch performs a PATCH request
func (c *Client) Patch(endpoint string, payload any) ([]byte, error) {
	return c.doRequest("PATCH", endpoint, nil, payload)
}

// Delete performs a DELETE request
func (c *Client) Delete(endpoint string, queryParams map[string]string) ([]byte, error) {
	return c.doRequest("DELETE", endpoint, queryParams, nil)
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

// doRequest performs the actual HTTP request
func (c *Client) doRequest(method, endpoint string, queryParams map[string]string, payload interface{}) ([]byte, error) {
	urlStr := fmt.Sprintf("%s%s/%s", c.BaseUrl, restApiPath, endpoint)
	if len(queryParams) > 0 {
		urlObj, err := url.Parse(urlStr)
		if err != nil {
			log.Printf("doRequest: failed to parse URL: %v", err)
			return nil, errors.New("failed to parse URL")
		}
		q := urlObj.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		urlObj.RawQuery = q.Encode()
		urlStr = urlObj.String()
	}

	var body io.Reader
	if payload != nil {
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("doRequest: failed to marshal payload: %v", err)
			return nil, errors.New("failed to marshal payload")
		}
		body = bytes.NewBuffer(jsonPayload)
	}

	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		log.Printf("doRequest: failed to create request: %v", err)
		return nil, errors.New("failed to create request")
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
		log.Printf("doRequest: failed to perform request: %v", err)
		return nil, errors.New("failed to perform request")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("doRequest: request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
		return nil, errors.New("request failed")
	}

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("doRequest: failed to read response body: %v", err)
		return nil, errors.New("failed to read response body")
	}

	return responseBody, nil
}
