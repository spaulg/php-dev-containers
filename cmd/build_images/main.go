package main

import (
	"fmt"
	"log"
	"main/internal/pkg/build"
)

func main() {
	err := RunBuild()
	if err != nil {
		log.Fatal(err)
	}
}

func RunBuild() error {
	// Parse command arguments to capture build information
	buildParameters := build.ParseArguments()

	// Connect Dagger client
	ctx, client := build.ConnectDaggerClient()
	defer client.Close()

	// Build package
	if _, err := buildOutput(buildParameters, ctx, client); err != nil {
		return fmt.Errorf("failed to build image: %w", err)
	}

	return nil
}
