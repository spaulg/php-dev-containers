package main

import (
	"context"
	"github.com/spaulg/php-dev-containers/internal/dagger"
	"fmt"
	"github.com/dchest/uniuri"
	"runtime"
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

func (m *PhpDevContainers) BuildPhpPackages(
	ctx context.Context,

// Source archive file path
	sourceArchive *dagger.File,

// List of architectures to build packages for, in addition to the native architecture
//+optional
	architectures *string,
) (*dagger.Directory, error) {
	var err error
	var container *dagger.Container
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

	// Download source archive
	sourceArchiveFileName := fmt.Sprintf("%s_%s.orig.tar.gz", m.PackageName, m.Version)

	// Start container
	container, err = dag.Container().
		From(m.BaseImage).
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

	// Prepare environment
	container, err = container.
		WithEnvVariable("DEBIAN_FRONTEND", "noninteractive").
		WithExec([]string{"apt", "update", "-y"}).
		WithExec([]string{"apt", "upgrade", "-y"}).
		WithExec([]string{"apt", "install", "-y", "build-essential", "devscripts", "quilt", "git", "sudo"}).
		WithExec([]string{"sh", "-c", "echo \"Cmnd_Alias DPKG_ADD_ARCH=/usr/bin/dpkg --add-architecture *\" >> /etc/sudoers.d/build"}).
		WithExec([]string{"sh", "-c", "echo \"Cmnd_Alias APT_INSTALL=/usr/bin/apt install -y *\" >> /etc/sudoers.d/build"}).
		WithExec([]string{"sh", "-c", "echo \"Cmnd_Alias APT_UPDATE=/usr/bin/apt update -y\" >> /etc/sudoers.d/build"}).
		WithExec([]string{"sh", "-c", "echo \"Cmnd_Alias APT_AUTOREMOVE=/usr/bin/apt autoremove * \" >> /etc/sudoers.d/build"}).
		WithExec([]string{"sh", "-c", "echo \"Cmnd_Alias MK_BUILD_DEPS=/usr/bin/mk-build-deps * \" >> /etc/sudoers.d/build"}).
		WithExec([]string{"sh", "-c", "echo \"build ALL=(ALL) NOPASSWD:MK_BUILD_DEPS, DPKG_ADD_ARCH, APT_INSTALL, APT_UPDATE, APT_AUTOREMOVE\" >> /etc/sudoers.d/build"}).
		WithExec([]string{"useradd", "-s", "/bin/bash", "-d", "/home/build", "-m", "-U", "build"}).
		WithWorkdir("/home/build").
		WithUser("build").
		WithExec([]string{"mkdir", "-p", "/home/build/packages"}).
		WithDirectory("/home/build/source", dag.CurrentModule().Source().Directory("assets/source/")).
		WithFile("/home/build/source/"+sourceArchiveFileName, sourceArchive).
		Sync(ctx)

	if err != nil {
		return nil, err
	}

	for architectureIndex, architecture := range buildArchitectures {
		if architectureIndex > 0 {
			// Install native packages for building phar in next architecture
			directory := container.Directory("/home/build/packages")
			files, err := directory.Glob(ctx, "**.deb")

			if err != nil {
				return nil, err
			}

			aptInstallCommand := []string{"sudo", "apt", "install", "-y", "--no-install-recommends", "--no-install-suggests"}
			for _, file := range files {
				if strings.HasSuffix(file, "_"+buildArchitectures[0]+".deb") || strings.HasSuffix(file, "_all.deb") {
					if strings.HasPrefix(file, "php"+m.ShortVersion+"-") {
						aptInstallCommand = append(aptInstallCommand, "/home/build/packages/"+file)
					}
				}
			}

			container, err = container.
				WithExec(aptInstallCommand).
				Sync(ctx)

			if err != nil {
				return nil, err
			}
		}

		container, err = container.
			WithWorkdir("/home/build").
			WithExec([]string{"rm", "-rf", m.BuildDirectoryPath}).
			WithExec([]string{"mkdir", "-p", m.BuildDirectoryPath}).
			WithWorkdir(m.BuildDirectoryPath).
			WithExec([]string{"cp", "/home/build/source/" + sourceArchiveFileName, m.BuildDirectoryRootPath + "/" + sourceArchiveFileName}).
			WithExec([]string{"tar", "-xzf", m.BuildDirectoryRootPath + "/" + sourceArchiveFileName, "--strip-components=1", "--exclude", "debian"}).
			WithExec([]string{"cp", "-R", "/home/build/source/" + m.ShortVersion, m.BuildDirectoryPath + "/debian"}).
			WithExec([]string{"rm", "-f", "debian/changelog"}).
			WithExec([]string{"debchange", "--create", "--package", m.PackageName, "--Distribution", "stable", "-v", m.Version + "-1php+dev+containers", m.Version + "-1php+dev+containers automated build"}).
			WithExec([]string{"make", "-f", "debian/rules", "prepare"}).
			WithExec([]string{"sudo", "dpkg", "--add-architecture", architecture}).
			WithExec([]string{"sudo", "apt", "update", "-y"}).
			WithExec([]string{"sudo", "mk-build-deps", "-i", "-t", "apt-get -o Debug::pkgProblemResolver=yes --no-install-recommends -y", "--host-arch", architecture}).
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
		buildDependenciesPackage := ""
		if architectureIndex == 0 {
			buildDependenciesPackage = m.PackageName + "-build-deps"
		} else {
			buildDependenciesPackage = m.PackageName + "-cross-build-deps"
		}

		container, err = container.
			WithExec([]string{"debuild", "-us", "-uc", "-a" + architecture}).
			WithExec([]string{"sudo", "apt", "autoremove", "-y", buildDependenciesPackage}).
			Sync(ctx)

		if err != nil {
			return nil, fmt.Errorf("failed to remove package build dependencies: %w", err)
		}
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
