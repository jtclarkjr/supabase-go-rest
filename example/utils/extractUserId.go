package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
)

func decodeTokenWithoutVerification(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("unable to decode token payload: %v", err)
	}

	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("unable to unmarshal token payload: %v", err)
	}

	return claims, nil
}

func ExtractUserId(tokenString string) (uuid.UUID, error) {
	log.Printf("Extracted token: %s", tokenString)

	claims, err := decodeTokenWithoutVerification(tokenString)
	if err != nil {
		log.Printf("Error decoding token: %v", err)
		return uuid.Nil, err
	}

	if userIDStr, ok := claims["sub"].(string); ok {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			log.Printf("Error parsing user ID as UUID: %v", err)
			return uuid.Nil, err
		}
		log.Printf("Extracted user ID from token: %s", userID)
		return userID, nil
	}
	log.Printf("User ID not found in token claims")
	return uuid.Nil, fmt.Errorf("user ID not found in token")
}
