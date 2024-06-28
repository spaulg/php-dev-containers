package main

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"log"
	"runtime"
	"unicode"
)

type PhpDevContainers struct {
	version      string // The major.minor.patch formatted version to build
	shortVersion string // The major.minor formatted version to build
	majorVersion string // The major version to build
	minorVersion string // The minor version to build
	patchVersion string // Patch version to build
	suffix       string // suffix to include in package name
	//	Architectures []string // Array of architectures to build
	packageName  string // Package name to build
	distribution string
	buildNumber  int

	sourceDirectory *Directory

	buildContainerImage    string
	buildDirectoryName     string
	buildDirectoryPath     string
	buildDirectoryRootPath string

	targetBaseContainerImage            string
	targetBuildContainerImageRepository string
	targetBuildContainerImageTag        string
	targetBuildContainerPlatform        Platform
	targetArchitecture                  string

	noCache         bool
	outputDirectory *Directory
}

const defaultDistribution = "bullseye"
const defaultBuildNumber = 1
const buildContainerImageRepository = "docker.io/spaulg/debuilder"
const targetBaseContainerImageRepository = "docker.io/debian"
const targetBuildContainerImageRepository = "docker.io/spaulg/php-dev-containers"
const packagePrefix = "php"
const packageDirectoryBase = "/home/build/packages"
const outputDirectoryBase = "assets/packages"

func New() *PhpDevContainers {
	return &PhpDevContainers{
		targetArchitecture:                  runtime.GOARCH,
		distribution:                        defaultDistribution,
		buildNumber:                         defaultBuildNumber,
		buildDirectoryRootPath:              packageDirectoryBase,
		targetBuildContainerImageRepository: targetBuildContainerImageRepository,
	}
}

func (m *PhpDevContainers) WithPhpVersion(version string) (*PhpDevContainers, error) {
	semanticVersion, err := semver.NewVersion(version)
	if err != nil {
		return m, fmt.Errorf("argument --version is not a valid semantic version: %v", err)
	}

	m.version = fmt.Sprintf("%d.%d.%d", semanticVersion.Major(), semanticVersion.Minor(), semanticVersion.Patch())
	m.shortVersion = fmt.Sprintf("%d.%d", semanticVersion.Major(), semanticVersion.Minor())
	m.majorVersion = fmt.Sprintf("%d", semanticVersion.Major())
	m.minorVersion = fmt.Sprintf("%d", semanticVersion.Minor())
	m.patchVersion = fmt.Sprintf("%d", semanticVersion.Patch())

	return m, nil
}

func (m *PhpDevContainers) WithPhpSuffix(suffix string) (*PhpDevContainers, error) {
	if suffix != "" {
		for _, r := range suffix {
			if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
				return m, fmt.Errorf("suffix must be alphanumeric")
			}
		}
	}

	m.suffix = "-" + suffix
	return m, nil
}

func (m *PhpDevContainers) WithSourceDirectory(sourceDirectory *Directory) *PhpDevContainers {
	m.sourceDirectory = sourceDirectory
	return m
}

func (m *PhpDevContainers) WithOutputDirectory(outputDirectory *Directory) *PhpDevContainers {
	m.outputDirectory = outputDirectory
	return m
}

//func (m *PhpDevContainers) WithPhpArchitectures(architectures string) *PhpDevContainers {
//	m.Architectures = strings.Split(architectures, ",")
//	return m
//}

func (m *PhpDevContainers) BuildPhpImage(ctx context.Context) (*Container, error) {
	// todo: validate that version, suffix, architectures have all been set

	m.packageName = packagePrefix + m.shortVersion + m.suffix

	// Complete derived fields
	m.buildContainerImage = buildContainerImageRepository + ":" + m.distribution
	m.buildDirectoryName = m.packageName + "_" + m.version
	m.buildDirectoryPath = m.buildDirectoryRootPath + "/" + m.buildDirectoryName
	m.targetBaseContainerImage = targetBaseContainerImageRepository + ":" + m.distribution

	m.targetBuildContainerImageRepository = targetBuildContainerImageRepository
	m.targetBuildContainerImageTag = m.packageName + "-" + m.targetArchitecture

	var err error
	if m.targetBuildContainerPlatform, err = mapContainerPlatform(m.targetArchitecture); err != nil {
		log.Fatal(err)
	}

	client := ConnectDaggerClient(ctx)
	defer client.Close()

	container, err := m.buildPackages(ctx)
	if err != nil {
		return container, fmt.Errorf("failed to build packages: %w", err)
	}

	return m.buildImage(ctx)
}

func mapContainerPlatform(targetPlatform string) (Platform, error) {
	switch targetPlatform {
	case "amd64":
		return "linux/amd64", nil
	case "arm64":
		return "linux/arm64", nil
	default:
		return "", fmt.Errorf("unsupport platform")
	}
}
