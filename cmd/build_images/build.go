package main

import (
	"context"
	"dagger.io/dagger"
	"github.com/dchest/uniuri"
	"main/internal/pkg/build"
)

func buildOutput(buildParameters *build.BuildParameters, ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	var container *dagger.Container

	// Start container
	container, err := client.Container(dagger.ContainerOpts{Platform: buildParameters.TargetBuildContainerPlatform}).
		From(buildParameters.TargetBaseContainerImage).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Bust cache if required
	if buildParameters.NoCache {
		container, err = container.
			WithEnvVariable("BURST_CACHE", uniuri.New()).
			Sync(ctx)

		if err != nil {
			return container, err
		}
	}

	container, err = container.
		WithDirectory("/packages", client.Host().Directory(buildParameters.OutputDirectoryPath)).
		WithExec([]string{"apt", "update", "-y"}).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Glob debian packages
	directory := container.Directory("/packages")
	files, err := directory.Glob(ctx, "**.deb")

	aptInstallCommand := []string{"apt", "install", "-y"}
	for _, file := range files {
		aptInstallCommand = append(aptInstallCommand, "/packages/"+file)
	}

	container, err = container.
		WithExec(aptInstallCommand).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Export container image by either saving locally or pushing to registry
	_, err = container.Publish(ctx, buildParameters.TargetBuildContainerImageRepository+":"+buildParameters.TargetBuildContainerImageTag)
	if err != nil {
		return container, err
	}

	return container, nil
}
