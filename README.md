# Supabase Go

Supabase Go Rest is a lightweight, flexible Go client designed to simplify interactions with Supabase's REST API, providing a seamless middleware solution for handling authenticated requests and Row Level Security (RLS) integrations.

## Features

### Authentication
- User Sign Up
- User Sign In
- Token Refresh
- Magic Link Authentication
- Password Recovery
- User Management

### Fluent Query Builder
- Method chaining API similar to supabase-js
- SELECT queries with column specification
- INSERT with automatic response handling
- UPDATE with filters
- DELETE with filters
- ORDER BY support
- LIMIT and SINGLE row queries
- Equality filters

### Advanced Capabilities
- Automatic Bearer token management
- Row Level Security (RLS) support
- Error handling for Supabase API interactions

## Installation

```bash
go get github.com/jtclarkjr/supabase-go-rest
```

## Quick Start

### Client Initialization

```go
import "github.com/jtclarkjr/supabase-go-rest"

supabaseUrl := "https://your-project.supabase.co"
supabaseKey := "your-supabase-api-key"
token := "optional-user-access-token"

client := supabase.NewClient(supabaseUrl, supabaseKey, token)
```

## Authentication Methods

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

## Fluent Query API

### SELECT Query

```go
// Get all rooms ordered by created_at
data, err := client.
    From("rooms").
    Select("*").
    Order("created_at", map[string]bool{"ascending": true}).
    Execute()
```

### SELECT with Filters

```go
// Get specific columns with filters
data, err := client.
    From("users").
    Select("id, name, email").
    Eq("age", "25").
    Limit(10).
    Execute()
```

### INSERT

```go
// Insert a new room and get it back
roomData := map[string]interface{}{
    "name": "My Room",
    "description": "A cool room",
}

newRoom, err := client.
    From("rooms").
    Insert(roomData).
    Select().
    Single().
    Execute()
```

### UPDATE

```go
// Update a room
updateData := map[string]interface{}{
    "name": "Updated Room Name",
}

updated, err := client.
    From("rooms").
    Update(updateData).
    Eq("id", "123").
    Select().
    Execute()
```

### DELETE

```go
// Delete a room
deleted, err := client.
    From("rooms").
    Delete().
    Eq("id", "123").
    Execute()
```

## Query Builder Methods

### From(table string)
Initiates a query on a specific table.

### Select(columns ...string)
Specifies columns to return. Use `"*"` or no arguments for all columns.

### Insert(data interface{})
Inserts data into the table. Automatically marshals to JSON.

### Update(data interface{})
Updates data in the table. Automatically marshals to JSON.

### Delete()
Marks the query as a delete operation.

### Eq(column, value string)
Adds an equality filter (`column = value`).

### Order(column string, opts map[string]bool)
Orders results by column. Options:
- `map[string]bool{"ascending": true}` - ascending order
- `map[string]bool{"ascending": false}` - descending order

### Limit(count int)
Limits the number of results.

### Single()
Expects a single row result (adds `limit=1`).

### Execute()
Executes the query and returns the response.

## Error Handling

The package provides custom error types:
- `ErrInvalidResponse`: Indicates an invalid server response
- `ErrRequestFailed`: Indicates a request failed to complete

```go
data, err := client.From("users").Select("*").Execute()
if err != nil {
    if errors.Is(err, supabase.ErrRequestFailed) {
        log.Printf("Request failed: %v", err)
    }
    return err
}
```

## Complete Example

```go
package main

import (
    "encoding/json"
    "fmt"
    "log"

    supabase "github.com/jtclarkjr/supabase-go-rest"
)

func main() {
    client := supabase.NewClient(
        "https://project.supabase.co",
        "api-key",
        "user-token",
    )
    
    // Get all users ordered by created_at
    usersData, err := client.
        From("users").
        Select("*").
        Order("created_at", map[string]bool{"ascending": true}).
        Execute()
    
    if err != nil {
        log.Fatalf("Failed to fetch users: %v", err)
    }
    
    var users []map[string]interface{}
    json.Unmarshal(usersData, &users)
    
    fmt.Printf("Found %d users\n", len(users))
    
    // Create a new user
    newUser := map[string]interface{}{
        "name":  "John Doe",
        "email": "john@example.com",
    }
    
    result, err := client.
        From("users").
        Insert(newUser).
        Select().
        Single().
        Execute()
    
    if err != nil {
        log.Fatalf("Failed to create user: %v", err)
    }
    
    fmt.Println("Created user:", string(result))
}
```

## API Interaction Examples

### Obtaining Authentication Token

```bash
curl -X POST http://localhost:8080/v1/auth/token \
-H "Content-Type: application/json" \
-d '{
  "email": "name@domain.com",
  "password": "somepassword"
}'
```

### Making Authenticated Requests

```bash
curl -X GET "https://localhost:8080/v1/food" \
 -H "Authorization: Bearer TOKEN_HERE"
```

### Practical Workflow

```bash
# Get token
TOKEN=$(curl -s -X POST http://localhost:8080/v1/auth/token \
-H "Content-Type: application/json" \
-d '{"email":"name@domain.com","password":"somepassword"}' | \
jq -r '.access_token')

# Use token
curl -X GET "https://localhost:8080/v1/food" \
 -H "Authorization: Bearer $TOKEN"
```

## Token Management

- Always include the `Authorization` header with a valid Bearer token
- Tokens are required for endpoints protected by Row Level Security (RLS)
- Tokens typically expire and need to be refreshed
- The client automatically handles Bearer token formatting

## PostgREST Query Operators

For advanced filtering, see [PostgREST operators documentation](https://docs.postgrest.org/en/v12/references/api/tables_views.html#operators).

## Related Projects

For more comprehensive Supabase functionality:
- [supabase-community/supabase-go](https://github.com/supabase-community/supabase-go)

## License

MIT License
