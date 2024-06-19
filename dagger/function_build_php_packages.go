package main

import (
	"context"
	"dagger/phpdevcontainers/internal/dagger"
	"fmt"
	"github.com/dchest/uniuri"
	"runtime"
	"strconv"
	"strings"
)

type PhpVersionAsset struct {
	Filename string `json:"filename"`
	Name     string `json:"name"`
}

type PhpVersion struct {
	Source []PhpVersionAsset `json:"source"`
	Museum bool              `json:"museum"`
}

func (m *PhpDevContainers) BuildPhpPackages(ctx context.Context, sourceArchive *dagger.File) (*dagger.Directory, error) {
	var err error
	var container *dagger.Container

	// Download source archive
	sourceArchiveFileName := fmt.Sprintf("%s_%s.orig.tar.gz", m.PackageName, m.Version)

	// Start container
	container, err = dag.Container().
		From(m.BuildContainerImage).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	// Bust cache if required
	if m.NoCache {
		container, err = container.
			WithEnvVariable("BURST_CACHE", uniuri.New()).
			Sync(ctx)

		if err != nil {
			return nil, err
		}
	}
	dag.Directory()

	container, err = container.
		WithDirectory("/home/build/source", dag.CurrentModule().Source().Directory("assets/source/")).
		WithFile("/home/build/source/"+sourceArchiveFileName, sourceArchive).
		WithExec([]string{"mkdir", "-p", m.BuildDirectoryPath}).
		WithWorkdir(m.BuildDirectoryPath).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	// Prepare package
	container, err = container.
		WithExec([]string{"cp", "/home/build/source/" + sourceArchiveFileName, m.BuildDirectoryRootPath + "/" + sourceArchiveFileName}).
		WithExec([]string{"tar", "-xzf", m.BuildDirectoryRootPath + "/" + sourceArchiveFileName, "--strip-components=1", "--exclude", "debian"}).
		WithExec([]string{"cp", "-R", "/home/build/source/" + m.ShortVersion, m.BuildDirectoryPath + "/debian"}).
		WithExec([]string{"rm", "-f", "debian/changelog"}).
		WithExec([]string{"debchange", "--create", "--package", m.PackageName, "--Distribution", "stable", "-v", m.Version + "-" + strconv.Itoa(m.BuildNumber), m.Version + "-" + strconv.Itoa(m.BuildNumber) + " automated build"}).
		WithExec([]string{"make", "-f", "debian/rules", "prepare"}).
		WithExec([]string{"sudo", "dpkg", "--add-architecture", runtime.GOARCH}).
		WithExec([]string{"sudo", "apt", "update", "-y"}).
		WithExec([]string{"sudo", "mk-build-deps", "-i", "-t", "apt-get -o Debug::pkgProblemResolver=yes --no-install-recommends -y", "--host-arch", runtime.GOARCH}).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	// Clean mk-build-deps files and delete
	buildDirectory := container.Directory(m.BuildDirectoryPath)
	var removeFiles []string

	for _, globPattern := range []string{"**.deb", "**.changes", "**.buildinfo"} {
		globFiles, err := buildDirectory.Glob(ctx, globPattern)

		if err != nil {
			return nil, fmt.Errorf("unable to list glob files for cleanup: %v", err)
		}

		for _, file := range globFiles {
			file = m.BuildDirectoryPath + "/" + file
			removeFiles = append(removeFiles, file)
		}
	}

	if len(removeFiles) > 0 {
		container, err = container.
			WithExec(append([]string{"rm", "-f"}, removeFiles...)).
			Sync(ctx)

		if err != nil {
			return nil, err
		}
	}

	// Final build
	container, err = container.
		WithExec([]string{"debuild", "-us", "-uc", "-a" + runtime.GOARCH}).
		Sync(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to build packages: %w", err)
	}

	directory := container.Directory("/home/build/packages/")
	entries, err := directory.Entries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list files from build: %w", err)
	}

	for _, file := range entries {
		if strings.HasSuffix(file, ".deb") == false {
			directory = directory.WithoutFile(file)
		}
	}

	return directory, nil
}
