# Supabase Go Rest

Supabase Go Client that makes use of Supabase Rest API. This only makes use of Supabase's Integration with Postgrest REST API.

Goal is to make use of Supabase REST API in Go to have a middle layer API between supabase while being able to handle tokens generated from supabase for Authenticated users to use RLS, such as tokens from client/app side.
This means the main requirement if using RLS on supabase side is needing to pass a token to the given request.

GET/POST/PUT/PATCH/DELETE operations

[Supabase REST API doc](https://supabase.com/docs/guides/api)

Alternatively use community package for other functionalies like storage and edge functions. [supabase-community/supabase-go](https://github.com/supabase-community/supabase-go)

## Examples

[example.go](https://github.com/jtclarkjr/supabase-go-rest/blob/main/example/example.go)

Example running in local for example.go code:
```
curl -X GET "https://localhost:8080/food" \
  -H "Authorization: Bearer TOKEN_HERE"
```
Here the point is you can define your endpoints and need to pass `Authorization` with Bearer token so that the supabase NewClient import can use the token.

