package main

import (
	"context"
	"dagger.io/dagger"
	"errors"
	"fmt"
	"log"
	"main/internal/pkg/build"
	"os"
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
	container, err := buildOutput(buildParameters, ctx, client)

	if err != nil {
		return fmt.Errorf("failed to build packages: %w", err)
	}

	if err := ExportArtifacts(buildParameters, ctx, container); err != nil {
		return fmt.Errorf("failed to export packages: %w", err)
	}

	return nil
}

func ExportArtifacts(buildParameters *build.BuildParameters, ctx context.Context, container *dagger.Container) error {
	// Create the output directory if it does not exist
	if _, err := os.Stat(buildParameters.OutputDirectoryPath); errors.Is(err, os.ErrNotExist) {
		log.Println("Creating export output directory: " + buildParameters.OutputDirectoryPath)
		if err = os.MkdirAll(buildParameters.OutputDirectoryPath, 0755); err != nil {
			return fmt.Errorf("failed to create export output directory: %w", err)
		}
	}

	// Export debian package files
	directory := container.Directory(buildParameters.BuildDirectoryRootPath)
	files, err := directory.Glob(ctx, "**.deb")

	if err != nil {
		return fmt.Errorf("encountered error whilst globbing files for export: %w", err)
	} else if len(files) > 0 {
		log.Println("Exporting files:")

		for _, file := range files {
			log.Println("  " + buildParameters.BuildDirectoryRootPath + "/" + file + " to " + buildParameters.OutputDirectoryPath + "/" + file)

			_, err = container.
				Directory(buildParameters.BuildDirectoryRootPath).
				File(file).
				Export(ctx, buildParameters.OutputDirectoryPath, dagger.FileExportOpts{AllowParentDirPath: true})

			if err != nil {
				return fmt.Errorf("failed to export file: %w", err)
			}
		}
	} else {
		log.Println("No matching files to export")
	}

	return nil
}
