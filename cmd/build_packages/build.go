package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"github.com/dchest/uniuri"
	"log"
	"main/internal/pkg/build"
	"strconv"
)

func buildOutput(buildParameters *build.BuildParameters, ctx context.Context, client *dagger.Client) (*dagger.Container, error) {
	var container *dagger.Container

	// Download source archive
	sourceArchiveFileName, err := downloadSourceArchive(buildParameters)

	if err != nil {
		return container, err
	}

	// Start container
	container, err = client.Container().
		From(buildParameters.BuildContainerImage).
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
		WithDirectory("/home/build/source", client.Host().Directory("assets/source")).
		WithExec([]string{"mkdir", "-p", buildParameters.BuildDirectoryPath}).
		WithWorkdir(buildParameters.BuildDirectoryPath).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Prepare package
	container, err = container.
		WithExec([]string{"cp", "/home/build/source/" + sourceArchiveFileName, buildParameters.BuildDirectoryRootPath + "/" + sourceArchiveFileName}).
		WithExec([]string{"tar", "-xzf", buildParameters.BuildDirectoryRootPath + "/" + sourceArchiveFileName, "--strip-components=1", "--exclude", "debian"}).
		WithExec([]string{"cp", "-R", "/home/build/source/" + buildParameters.ShortVersion, buildParameters.BuildDirectoryPath + "/debian"}).
		WithExec([]string{"rm", "-f", "debian/changelog"}).
		WithExec([]string{"debchange", "--create", "--package", buildParameters.PackageName, "--distribution", "stable", "-v", buildParameters.Version + "-" + strconv.Itoa(buildParameters.BuildNumber), buildParameters.Version + "-" + strconv.Itoa(buildParameters.BuildNumber) + " automated build"}).
		WithExec([]string{"make", "-f", "debian/rules", "prepare"}).
		WithExec([]string{"sudo", "dpkg", "--add-architecture", buildParameters.Architecture}).
		WithExec([]string{"sudo", "apt", "update", "-y"}).
		WithExec([]string{"sudo", "mk-build-deps", "-i", "-t", "apt-get -o Debug::pkgProblemResolver=yes --no-install-recommends -y", "--host-arch", buildParameters.Architecture}).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Clean mk-build-deps files and delete
	buildDirectory := container.Directory(buildParameters.BuildDirectoryPath)
	var removeFiles []string

	for _, globPattern := range []string{"**.deb", "**.changes", "**.buildinfo"} {
		globFiles, err := buildDirectory.Glob(ctx, globPattern)

		if err != nil {
			return container, fmt.Errorf("unable to list glob files for cleanup: %v", err)
		}

		for _, file := range globFiles {
			file = buildParameters.BuildDirectoryPath + "/" + file

			log.Println("Removing file: " + file)
			removeFiles = append(removeFiles, file)
		}
	}

	if len(removeFiles) > 0 {
		container, err = container.
			WithExec(append([]string{"rm", "-f"}, removeFiles...)).
			Sync(ctx)

		if err != nil {
			return container, err
		}
	}

	// Final build
	return container.
		WithExec([]string{"debuild", "-us", "-uc", "-a" + buildParameters.Architecture}).
		Sync(ctx)
}
