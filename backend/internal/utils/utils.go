package utils

import "github.com/google/uuid"

func GenerateClientID() string {
	return "app-" + uuid.New().String()
}
