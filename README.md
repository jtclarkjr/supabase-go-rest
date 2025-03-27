# Supabase Go Rest

## 🚀 Overview

Supabase Go Rest is a lightweight, flexible Go client designed to simplify interactions with Supabase's REST API, providing a seamless middleware solution for handling authenticated requests and Row Level Security (RLS) integrations.

## 📦 Features

### Authentication
- User Sign Up
- User Sign In
- Token Refresh
- Magic Link Authentication
- Password Recovery
- User Management

### REST API Methods
- GET
- POST
- PUT
- PATCH
- DELETE

### Advanced Capabilities
- Automatic Bearer token management
- Row Level Security (RLS) support
- Flexible query parameter handling
- Error handling for Supabase API interactions

## 🛠 Installation

```bash
go get github.com/jtclarkjr/supabase-go-rest
```

## 🔧 Quick Start

### Client Initialization

```go
import "github.com/jtclarkjr/supabase-go-rest"

supabaseUrl := "https://your-project.supabase.co"
supabaseKey := "your-supabase-api-key"
token := "optional-user-access-token"

client := supabase.NewClient(supabaseUrl, supabaseKey, token)
```

## 🔐 Authentication Methods

### Sign Up
```go
body, err := client.SignUp("user@example.com", "password")
```

### Sign In
```go
authResponse, err := client.SignIn("user@example.com", "password")
```

### Refresh Token
```go
newTokenResponse, err := client.RefreshToken(refreshToken)
```

### Send Magic Link
```go
body, err := client.SendMagicLink("user@example.com")
```

### Password Recovery
```go
body, err := client.SendPasswordRecovery("user@example.com")
```

## 🌐 REST API Interactions

### GET Request
```go
// Simple GET request with query parameters
queryParams := map[string]string{
    "name": "eq.John",
    "age": "gt.25"
}
body, err := client.Get("Users", queryParams)
```

### POST Request
```go
data := map[string]interface{}{
    "name": "John Doe",
    "email": "john@example.com"
}
jsonData, _ := json.Marshal(data)
body, err := client.Post("Users", jsonData)
```

### PUT Request
```go
data := map[string]interface{}{
    "name": "Updated Name"
}
jsonData, _ := json.Marshal(data)
body, err := client.Put("Users", "id", "123", jsonData)
```

### PATCH Request
```go
data := map[string]interface{}{
    "last_login": "2024-03-28"
}
jsonData, _ := json.Marshal(data)
queryParams := map[string]string{"email": "john@example.com"}
body, err := client.Patch("Users", queryParams, jsonData)
```

### DELETE Request
```go
body, err := client.Delete("Users", "id", "123")
```

## 🔍 Query Parameter Operators

Supabase Go Rest supports PostgREST query operators for advanced
[filtering](https://docs.postgrest.org/en/v12/references/api/tables_views.html#operators)
## 🚨 Error Handling

The package provides custom error types:

- `ErrInvalidResponse`: Indicates an invalid server response
- `ErrRequestFailed`: Indicates a request failed to complete

## 📡 Example Use Case
[example.go](https://github.com/jtclarkjr/supabase-go-rest/blob/main/example/example.go)

```go
package main

import (
    "fmt"
    "log"

    supabase "github.com/jtclarkjr/supabase-go-rest"
)

func main() {
    client := supabase.NewClient("https://project.supabase.co", "api-key", "user-token")
    
    // Fetch users over 25
    body, err := client.Get("Users", map[string]string{
        "age": "gt.25"
    })
    
    if err != nil {
        log.Fatalf("Failed to fetch users: %v", err)
    }
    
    fmt.Println(string(body))
}
```

## 🌐 API Interaction Examples

### Obtaining Authentication Token

To get an authentication token, use the following cURL command:

```bash
curl -X POST http://localhost:8080/v1/auth/token \
-H "Content-Type: application/json" \
-d '{
 "email": "name@domain.com",
 "password": "somepassword"
}'
```

This will return a JSON response with an access token.

### Making Authenticated Requests

Once you have the token, use it in the Authorization header for subsequent requests:

```bash
curl -X GET "https://localhost:8080/v1/food" \
 -H "Authorization: Bearer TOKEN_HERE"
```

### Practical Example Workflow

1. Get Authentication Token:
```bash
# Request token
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/token \
-H "Content-Type: application/json" \
-d '{"email":"name@domain.com","password":"somepassword"}' | \
jq -r '.access_token')

# Use token in subsequent request
curl -X GET "https://localhost:8080/v1/food" \
 -H "Authorization: Bearer $TOKEN"
```

## 🔑 Token Management Notes

- Always include the `Authorization` header with a valid Bearer token
- Tokens are required for endpoints protected by Row Level Security (RLS)
- Tokens typically expire and need to be refreshed
- The Supabase client automatically handles token formatting

[... rest of the previous content remains the same ...]

## 📝 Important Notes

- Always provide an access token for authenticated requests
- Automatic Bearer token formatting
- Support for both manual and Supabase-generated tokens

## 🔗 Related Projects

For more comprehensive Supabase functionality:
- [supabase-community/supabase-go](https://github.com/supabase-community/supabase-go)

## 📄 License

MIT License

