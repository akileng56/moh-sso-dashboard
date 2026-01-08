package utils

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func GenerateClientID() string {
	return "app-" + uuid.New().String()
}

// helper
func ToNullUUID(s string) uuid.NullUUID {
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

func ProjectRoot() string {
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Current working directory: %s", wd)

	for {
		if _, err := os.Stat(filepath.Join(wd, "app.env")); err == nil {
			return wd
		}

		parent := filepath.Dir(wd)
		if parent == wd {
			log.Fatal("app.env not found in any parent directory")
		}
		wd = parent
	}
}

func NormalizeHeader(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

func SplitClientIDs(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		v := strings.TrimSpace(p)
		if v != "" {
			out = append(out, v)
		}
	}
	return out
}

func ParseBoolDefaultTrue(s string) (bool, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return true, nil
	}
	switch s {
	case "true", "1", "yes", "y":
		return true, nil
	case "false", "0", "no", "n":
		return false, nil
	default:
		return false, fmt.Errorf("invalid enabled value: %q", s)
	}
}

func EscapeCSV(s string) string {
	s = strings.ReplaceAll(s, `"`, `""`)
	if strings.ContainsAny(s, ",\n\r") {
		return `"` + s + `"`
	}
	return s
}

func TokenHasRealmRole(tokenStr string, role string) bool {
	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return false
	}

	claims := token.Claims.(jwt.MapClaims)

	ra, ok := claims["realm_access"].(map[string]interface{})
	if !ok {
		return false
	}

	roles, ok := ra["roles"].([]interface{})
	if !ok {
		return false
	}

	for _, r := range roles {
		if r.(string) == role {
			return true
		}
	}

	return false
}

func MustJSON(v interface{}) json.RawMessage {
	b, _ := json.Marshal(v)
	return b
}
