package main

import (
	"context"
	"fmt"
	"github.com/dchest/uniuri"
	"log"
	"strconv"
)

func (m *PhpDevContainers) buildPackages(ctx context.Context) (*Container, error) {
	var container *Container

	// Download source archive
	sourceArchiveFileName, err := m.downloadSourceArchive()

	if err != nil {
		return container, err
	}

	// Start container
	container, err = dag.Container().
		From(m.buildContainerImage).
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
	dag.Directory()

	container, err = container.
		WithDirectory("/home/build/source", m.sourceDirectory).
		WithExec([]string{"mkdir", "-p", m.buildDirectoryPath}).
		WithWorkdir(m.buildDirectoryPath).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Prepare package
	container, err = container.
		WithExec([]string{"cp", "/home/build/source/" + sourceArchiveFileName, m.buildDirectoryRootPath + "/" + sourceArchiveFileName}).
		WithExec([]string{"tar", "-xzf", m.buildDirectoryRootPath + "/" + sourceArchiveFileName, "--strip-components=1", "--exclude", "debian"}).
		WithExec([]string{"cp", "-R", "/home/build/source/" + m.shortVersion, m.buildDirectoryPath + "/debian"}).
		WithExec([]string{"rm", "-f", "debian/changelog"}).
		WithExec([]string{"debchange", "--create", "--package", m.packageName, "--distribution", "stable", "-v", m.version + "-" + strconv.Itoa(m.buildNumber), m.version + "-" + strconv.Itoa(m.buildNumber) + " automated build"}).
		WithExec([]string{"make", "-f", "debian/rules", "prepare"}).
		WithExec([]string{"sudo", "dpkg", "--add-architecture", m.targetArchitecture}).
		WithExec([]string{"sudo", "apt", "update", "-y"}).
		WithExec([]string{"sudo", "mk-build-deps", "-i", "-t", "apt-get -o Debug::pkgProblemResolver=yes --no-install-recommends -y", "--host-arch", m.targetArchitecture}).
		Sync(ctx)

	if err != nil {
		return container, err
	}

	// Clean mk-build-deps files and delete
	buildDirectory := container.Directory(m.buildDirectoryPath)
	var removeFiles []string

	for _, globPattern := range []string{"**.deb", "**.changes", "**.buildinfo"} {
		globFiles, err := buildDirectory.Glob(ctx, globPattern)

		if err != nil {
			return container, fmt.Errorf("unable to list glob files for cleanup: %v", err)
		}

		for _, file := range globFiles {
			file = m.buildDirectoryPath + "/" + file

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
		WithExec([]string{"debuild", "-us", "-uc", "-a" + m.targetArchitecture}).
		Sync(ctx)
}
