package main

import (
	"context"
	"dagger/phpdevcontainers/internal/dagger"
	"dagger/phpdevcontainers/utils"
	"github.com/dchest/uniuri"
	"runtime"
	"strings"
)

const DockerRepository = "index.docker.io/spaulg/php-dev-containers"

func (m *PhpDevContainers) BuildPhpImage(
	ctx context.Context,

	// Packages directory path
	packageDirectory *dagger.Directory,

	// List of architectures to build packages for, in addition to the native architecture
	//+optional
	architectures *string,

	// Push all container platform variants built under a single container manifest
	//+optional
	push bool,
) error {
	var buildArchitectures []string

	// Process architecture list, ensuring the current runtime arch is first
	buildArchitectures = append(buildArchitectures, runtime.GOARCH)

	if architectures != nil {
		architectureList := strings.Split(*architectures, ",")
		for _, architecture := range architectureList {
			if architecture != runtime.GOARCH {
				buildArchitectures = append(buildArchitectures, strings.TrimSpace(architecture))
			}
		}
	}

	platformVariants := make([]*dagger.Container, 0, len(buildArchitectures))
	for _, architecture := range buildArchitectures {
		platform, err := utils.MapContainerPlatform(architecture)

		if err != nil {
			return err
		}

		// Start container
		container, err := dag.Container(dagger.ContainerOpts{Platform: platform}).
			From(m.BaseImage).
			Sync(ctx)

		if err != nil {
			return err
		}

		// Bust cache if required
		if m.NoCache {
			container, err = container.
				WithEnvVariable("BURST_CACHE", uniuri.New()).
				Sync(ctx)

			if err != nil {
				return err
			}
		}

		container, err = container.
			WithMountedDirectory("/packages", packageDirectory).
			WithExec([]string{"sh", "-c", "rm /var/lib/dpkg/info/libc-bin.*"}).
			WithExec([]string{"apt-get", "clean"}).
			WithExec([]string{"apt", "update", "-y"}).
			WithExec([]string{"apt-get", "install", "libc-bin"}).
			Sync(ctx)

		if err != nil {
			return err
		}

		// Glob debian packages
		directory := container.Directory("/packages")
		files, err := directory.Glob(ctx, "**.deb")

		if err != nil {
			return err
		}

		aptInstallCommand := []string{"apt", "install", "-y", "--no-install-recommends", "--no-install-suggests"}
		for _, file := range files {
			if strings.HasSuffix(file, "_"+architecture+".deb") || strings.HasSuffix(file, "_all.deb") {
				aptInstallCommand = append(aptInstallCommand, "/packages/"+file)
			}
		}

		container, err = container.
			WithExec(aptInstallCommand).
			WithExec([]string{"sh", "-c", "dpkg -l | grep \"1php+dev+containers\" | awk '{print $2}' | xargs apt-mark hold"}).
			WithExec([]string{"apt", "install", "-y", "build-essential", "devscripts", "quilt", "git"}).
			Sync(ctx)

		if err != nil {
			return err
		}

		platformVariants = append(platformVariants, container)
	}

	if push {
		var err error
		_, err = dag.Container().Publish(ctx, DockerRepository+":"+m.TagName, dagger.ContainerPublishOpts{
			PlatformVariants: platformVariants,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
