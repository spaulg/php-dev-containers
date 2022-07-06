package main

import (
	"log"
	"main/internal/pkg/build"
	"os"
)

func main() {
	os.Exit(RunBuild())
}

func RunBuild() int {
	// Parse command arguments to capture build information
	buildParameters := build.ParseArguments()

	// Connect Dagger client
	ctx, client := build.ConnectDaggerClient()
	defer client.Close()

	// Build package
	_, err := buildOutput(buildParameters, ctx, client)

	if err != nil {
		log.Println(err)
	}

	return 0
}
