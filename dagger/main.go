package main

import (
    "context"
    "dagger.io/dagger"
    "fmt"
    "github.com/Masterminds/semver/v3"
    "strings"
    "unicode"
)

type PhpDevContainers struct {
    Version       string   // The major.minor.patch formatted version to build
    ShortVersion  string   // The major.minor formatted version to build
    MajorVersion  string   // The major version to build
    MinorVersion  string   // The minor version to build
    PatchVersion  string   // Patch version to build
    Suffix        string   // Suffix to include in package name
    Architectures []string // Array of architectures to build
    PackageName   string   // Package name to build

    BuildContainerImage    string
    BuildDirectoryName     string
    BuildDirectoryPath     string
    BuildDirectoryRootPath string

    TargetBaseContainerImage            string
    TargetBuildContainerImageRepository string
    TargetBuildContainerImageTag        string
    TargetBuildContainerPlatform        dagger.Platform
    TargetArchitecture                  string
}

const packagePrefix = "php"

func (m *PhpDevContainers) WithPhpVersion(version string) (*PhpDevContainers, error) {
    semanticVersion, err := semver.NewVersion(version)
    if err != nil {
        return m, fmt.Errorf("argument --version is not a valid semantic version: %v", err)
    }

    m.Version = fmt.Sprintf("%d.%d.%d", semanticVersion.Major(), semanticVersion.Minor(), semanticVersion.Patch())
    m.ShortVersion = fmt.Sprintf("%d.%d", semanticVersion.Major(), semanticVersion.Minor())
    m.MajorVersion = fmt.Sprintf("%d", semanticVersion.Major())
    m.MinorVersion = fmt.Sprintf("%d", semanticVersion.Minor())
    m.PatchVersion = fmt.Sprintf("%d", semanticVersion.Patch())

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

    m.Suffix = "-" + suffix
    return m, nil
}

func (m *PhpDevContainers) WithPhpArchitectures(architectures string) *PhpDevContainers {
    m.Architectures = strings.Split(architectures, ",")
    return m
}

func (m *PhpDevContainers) BuildPhpImage(ctx context.Context) (*dagger.Container, error) {
    // todo: validate that version, suffix, architectures have all been set

    m.PackageName = packagePrefix + m.ShortVersion + m.Suffix

    // todo: set other parameters required for build

    client := ConnectDaggerClient(ctx)
    defer client.Close()

    container, err := m.buildPackages(ctx, client)
    if err != nil {
        return container, fmt.Errorf("failed to build packages: %w", err)
    }

    return m.buildImage(ctx, client)
}
