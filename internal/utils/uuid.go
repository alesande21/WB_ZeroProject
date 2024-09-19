package utils

import (
	"github.com/google/uuid"
	"time"
)

func GenerateUUID() string {
	return uuid.New().String()
}

func GenerateUUIDV7() (string, error) {
	uuidNew, err := uuid.NewV7()
	return uuidNew.String(), err
}

func GetCurrentTimeRFC3339() string {
	currentTime := time.Now()
	formattedTime := currentTime.Format(time.RFC3339)
	return formattedTime
}
