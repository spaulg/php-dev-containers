package utils

import (
	"dagger/phpdevcontainers/internal/dagger"
	"fmt"
)

func MapContainerPlatform(targetPlatform string) (dagger.Platform, error) {
	switch targetPlatform {
	case "amd64":
		return "linux/amd64", nil
	case "arm64":
		return "linux/arm64", nil
	default:
		return "", fmt.Errorf("unsupported platform: %v", targetPlatform)
	}
}
