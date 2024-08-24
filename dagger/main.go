package main

import (
	"fmt"
	"github.com/Masterminds/semver"
	"slices"
	"strings"
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

	// Package name to build
	// +private
	PackageName string

	// +private
	BaseImage string

	// +private
	TagName string

	// +private
	BuildDirectoryPath string

	// +private
	BuildDirectoryRootPath string

	// +private
	NoCache bool
}

const baseImage = "docker.io/debian"
const packagePrefix = "php"
const packageDirectoryBase = "/home/build/packages"

func New(
	// Version to build
	version string,

	// Debian distribution
	distribution string,

	// Burst cache
	// +optional
	burstCache bool,
) (*PhpDevContainers, error) {
	semanticVersion, err := semver.NewVersion(version)
	if err != nil {
		return nil, fmt.Errorf("argument --Version is not a valid semantic Version: %v", err)
	}

	suffix := ""
	metadata := semanticVersion.Metadata()
	metadataList := strings.Split(metadata, ".")
	if slices.Contains(metadataList, "zts") {
		suffix = "-zts"
	}

	fullVersion := fmt.Sprintf("%d.%d.%d", semanticVersion.Major(), semanticVersion.Minor(), semanticVersion.Patch())
	shortVersion := fmt.Sprintf("%d.%d", semanticVersion.Major(), semanticVersion.Minor())
	majorVersion := fmt.Sprintf("%d", semanticVersion.Major())
	minorVersion := fmt.Sprintf("%d", semanticVersion.Minor())
	patchVersion := fmt.Sprintf("%d", semanticVersion.Patch())

	packageName := packagePrefix + shortVersion + suffix
	tagName := shortVersion + suffix

	// Complete derived fields
	buildDirectoryPath := packageDirectoryBase + "/" + packageName + "_" + fullVersion

	return &PhpDevContainers{
		Version:      fullVersion,
		ShortVersion: shortVersion,
		MajorVersion: majorVersion,
		MinorVersion: minorVersion,
		PatchVersion: patchVersion,
		Suffix:       suffix,
		PackageName:  packageName,
		NoCache:      burstCache,

		BaseImage: baseImage + ":" + distribution,
		TagName:   tagName,

		BuildDirectoryRootPath: packageDirectoryBase,
		BuildDirectoryPath:     buildDirectoryPath,
	}, nil
}
