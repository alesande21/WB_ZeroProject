package utils

import (
	"github.com/google/uuid"
)

func GenerateUUIDV7() (string, error) {
	uuidNew, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	return uuidNew.String(), err
}
