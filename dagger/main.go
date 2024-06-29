package main

import (
	"context"
	"fmt"
	"github.com/Masterminds/semver"
	"runtime"
	"unicode"
)

type PhpDevContainers struct {
	// The major.minor.patch formatted Version to build
	// +private
	Version string

	// The major.minor formatted Version to build
	// +private
	ShortVersion string

	// The major Version to build
	// +private
	MajorVersion string

	// The minor Version to build
	// +private
	MinorVersion string

	// Patch Version to build
	// +private
	PatchVersion string

	// Suffix to include in package name
	// +private
	Suffix string

	// Array of architectures to build
	// +private
	//	Architectures []string

	// Package name to build
	// +private
	PackageName string

	// +private
	Distribution string

	// +private
	BuildNumber int

	// +private
	SourceDirectory *Directory

	// +private
	BuildContainerImage string

	// +private
	BuildDirectoryName string

	// +private
	BuildDirectoryPath string

	// +private
	BuildDirectoryRootPath string

	// +private
	TargetBaseContainerImage string

	// +private
	TargetBuildContainerImageRepository string

	// +private
	TargetBuildContainerImageTag string

	// +private
	TargetBuildContainerPlatform Platform

	// +private
	TargetArchitecture string

	// +private
	NoCache bool

	// +private
	OutputDirectory *Directory
}

const defaultDistribution = "bullseye"
const defaultBuildNumber = 1
const buildContainerImageRepository = "docker.io/spaulg/debuilder"
const targetBaseContainerImageRepository = "docker.io/debian"
const targetBuildContainerImageRepository = "docker.io/spaulg/php-dev-containers"
const packagePrefix = "php"
const packageDirectoryBase = "/home/build/packages"
const outputDirectoryBase = "assets/packages"

func New(version string, sourceDirectory *Directory, outputDirectory *Directory) (*PhpDevContainers, error) {
	semanticVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("argument --Version is not a valid semantic Version: %v", err)
	}

	return &PhpDevContainers{
		TargetArchitecture:                  runtime.GOARCH,
		Distribution:                        defaultDistribution,
		BuildNumber:                         defaultBuildNumber,
		BuildDirectoryRootPath:              packageDirectoryBase,
		TargetBuildContainerImageRepository: targetBuildContainerImageRepository,

		Version:      fmt.Sprintf("%d.%d.%d", semanticVersion.Major(), semanticVersion.Minor(), semanticVersion.Patch()),
		ShortVersion: fmt.Sprintf("%d.%d", semanticVersion.Major(), semanticVersion.Minor()),
		MajorVersion: fmt.Sprintf("%d", semanticVersion.Major()),
		MinorVersion: fmt.Sprintf("%d", semanticVersion.Minor()),
		PatchVersion: fmt.Sprintf("%d", semanticVersion.Patch()),

		SourceDirectory: sourceDirectory,
		OutputDirectory: outputDirectory,
	}, nil
}

func (m *PhpDevContainers) WithPhpSuffix(suffix string) (*PhpDevContainers, error) {
	if suffix != "" {
		for _, r := range suffix {
			if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
				return m, fmt.Errorf("Suffix must be alphanumeric")
			}
		}
	}

	m.Suffix = "-" + suffix
	return m, nil
}

//func (m *PhpDevContainers) WithPhpArchitectures(architectures string) *PhpDevContainers {
//	m.Architectures = strings.Split(architectures, ",")
//	return m
//}

func (m *PhpDevContainers) BuildPhpImage(ctx context.Context) (*Container, error) {
	// todo: validate that Version, Suffix, architectures have all been set

	m.PackageName = packagePrefix + m.ShortVersion + m.Suffix

	// Complete derived fields
	m.BuildContainerImage = buildContainerImageRepository + ":" + m.Distribution
	m.BuildDirectoryName = m.PackageName + "_" + m.Version
	m.BuildDirectoryPath = m.BuildDirectoryRootPath + "/" + m.BuildDirectoryName
	m.TargetBaseContainerImage = targetBaseContainerImageRepository + ":" + m.Distribution

	m.TargetBuildContainerImageRepository = targetBuildContainerImageRepository
	m.TargetBuildContainerImageTag = m.PackageName + "-" + m.TargetArchitecture

	var err error
	if m.TargetBuildContainerPlatform, err = mapContainerPlatform(m.TargetArchitecture); err != nil {
		return nil, err
	}

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
		return "", fmt.Errorf("unsupported platform: %v", targetPlatform)
	}
}
