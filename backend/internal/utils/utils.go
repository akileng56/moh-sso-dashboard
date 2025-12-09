package utils

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/google/uuid"
)

func GenerateClientID() string {
	return "app-" + uuid.New().String()
}

// helper
func ToNullUUID(s string) uuid.NullUUID { // <-- ADDED
	if s == "" {
		return uuid.NullUUID{Valid: false}
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: id, Valid: true}
}

func ExtractUserIDFromJWT(token string) string {
	if token == "" {
		return ""
	}

	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return ""
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(payload, &data); err != nil {
		return ""
	}

	// Keycloak stores user ID in "sub"
	if sub, ok := data["sub"].(string); ok {
		return sub
	}

	return ""
}

func Encode(v interface{}) []byte {
	if v == nil {
		return nil
	}
	b, _ := json.Marshal(v)
	return b
}
