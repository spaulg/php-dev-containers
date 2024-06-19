package main

import (
	"context"
	"dagger.io/dagger"
)

type PhpDevContainers struct{}

func (m *PhpDevContainers) BuildPhp(ctx context.Context, version string, suffix *string) *dagger.Container {
	dag, _ := dagger.Connect(ctx)
	container := dag.Container().
		From("alpine:latest").
		WithExec([]string{"sh", "-c", "echo 'version: " + version + "' >> /debug"})

	if suffix != nil {
		container = container.
			WithExec([]string{"sh", "-c", "echo 'suffix: " + *suffix + "' >> /debug"})
	}

	return container
}
