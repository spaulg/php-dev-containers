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
	if _, err := os.Stat(buildParameters.OutputDirectoryPath); errors.Is(err, os.ErrNotExist) {
		log.Println("Creating export output directory: " + buildParameters.OutputDirectoryPath)
		if err = os.MkdirAll(buildParameters.OutputDirectoryPath, 0755); err != nil {
			log.Println("Failed to create export output directory")
			log.Println(err)
			return 1
		}
	}

	// Export debian package files
	exitCode := 0
	directory := container.Directory(buildParameters.BuildDirectoryRootPath)
	files, err := directory.Glob(ctx, "**.deb")

	if err != nil {
		log.Println("Encountered error whilst globbing files for export")
		exitCode = 1
	} else if len(files) > 0 {
		log.Println("Exporting files:")

		for _, file := range files {
			log.Println("  " + buildParameters.BuildDirectoryRootPath + "/" + file + " to " + buildParameters.OutputDirectoryPath + "/" + file)

			_, err = container.
				Directory(buildParameters.BuildDirectoryRootPath).
				File(file).
				Export(ctx, buildParameters.OutputDirectoryPath, dagger.FileExportOpts{AllowParentDirPath: true})

			if err != nil {
				log.Println(err)
				exitCode = 1
			}
		}
	} else {
		log.Println("No matching files to export")
	}

	return exitCode
}
