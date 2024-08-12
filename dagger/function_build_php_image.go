package main

import (
	"context"
	"dagger/phpdevcontainers/internal/dagger"
	"github.com/dchest/uniuri"
	"runtime"
	"strings"
)

func (m *PhpDevContainers) BuildPhpImage(ctx context.Context, packageDirectory *dagger.Directory) (*dagger.Container, error) {
	var container *dagger.Container

	// Start container
	container, err := dag.Container().
		From("docker.io/debian:bullseye").
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Bust cache if required
	if m.NoCache {
		container, err = container.
			WithEnvVariable("BURST_CACHE", uniuri.New()).
			Sync(ctx)

		if err != nil {
			return container, err
		}
	}

	container, err = container.
		WithDirectory("/packages", packageDirectory).
		WithExec([]string{"apt", "update", "-y"}).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Glob debian packages
	directory := container.Directory("/packages")
	files, err := directory.Glob(ctx, "**.deb")

	if err != nil {
		return container, err
	}

	aptInstallCommand := []string{"apt", "install", "-y"}
	for _, file := range files {
		if strings.HasSuffix(file, "_"+runtime.GOARCH+".deb") || strings.HasSuffix(file, "_all.deb") {
			aptInstallCommand = append(aptInstallCommand, "/packages/"+file)
		}
	}

	return container.
		WithExec(aptInstallCommand).
		Sync(ctx)
}
