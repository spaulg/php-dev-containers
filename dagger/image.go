package main

import (
	"context"
	"github.com/dchest/uniuri"
)

func (m *PhpDevContainers) buildImage(ctx context.Context) (*Container, error) {
	var container *Container

	// Start container
	container, err := dag.Container(ContainerOpts{Platform: m.targetBuildContainerPlatform}).
		From(m.targetBaseContainerImage).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Bust cache if required
	if m.noCache {
		container, err = container.
			WithEnvVariable("BURST_CACHE", uniuri.New()).
			Sync(ctx)

		if err != nil {
			return container, err
		}
	}

	container, err = container.
		WithDirectory("/packages", m.outputDirectory).
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
	_, err = container.Publish(ctx, m.targetBuildContainerImageRepository+":"+m.targetBuildContainerImageTag)
	if err != nil {
		return container, err
	}

	return container, nil
}
