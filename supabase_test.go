package supabase

import "testing"

func TestNewClient(t *testing.T) {
	baseUrl := "https://example.supabase.co"
	apiKey := "your_api_key"
	token := "your_token"

	client := NewClient(baseUrl, apiKey, token)

	if client.BaseUrl != baseUrl {
		t.Errorf("Expected BaseUrl to be %s, got %s", baseUrl, client.BaseUrl)
	}
	if client.ApiKey != apiKey {
		t.Errorf("Expected ApiKey to be %s, got %s", apiKey, client.ApiKey)
	}
	if client.Token != token {
		t.Errorf("Expected Token to be %s, got %s", token, client.Token)
	}
}
