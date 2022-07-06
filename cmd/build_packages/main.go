package main

import (
	"context"
	"dagger.io/dagger"
	"errors"
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
	container, err := buildOutput(buildParameters, ctx, client)

	if err != nil {
		log.Println(err)
	}

	return ExportArtifacts(buildParameters, ctx, container)
}

func ExportArtifacts(buildParameters *build.BuildParameters, ctx context.Context, container *dagger.Container) int {
	// Create the output directory if it does not exist
	if _, err := os.Stat("assets/packages"); errors.Is(err, os.ErrNotExist) {
		if err = os.MkdirAll("assets/packages", 0755); err != nil {
			log.Println(err)
			return 1
		}
	}

	// Export debian package files
	exitCode := 0
	directory := container.Directory("/home/build/packages")
	files, err := directory.Glob(ctx, "**.deb")

	if err != nil {
		exitCode = 1
	} else {
		for _, file := range files {
			_, err = container.
				Directory("/home/build/packages").
				File(file).
				Export(ctx, "assets/packages", dagger.FileExportOpts{AllowParentDirPath: true})

			if err != nil {
				log.Println(err)
				exitCode = 1
			}
		}
	}

	return exitCode
}
