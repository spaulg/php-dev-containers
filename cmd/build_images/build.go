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
	container, err := client.Container().
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

	//container, err = container.
	//	WithDirectory("/packages", client.Host().Directory("assets/packages")).
	//	WithExec([]string{"apt", "install", "/packages/*.deb"}).
	//	Sync(ctx)
	//
	//if err != nil {
	//	return container, err
	//}
	//
	// Export container image by either saving locally or pushing to registry
	// _, err = container.Publish(ctx, buildParameters.TargetBuildContainerImage)
	// if err != nil {
	//	return container, err
	// }

	return container, nil
}
