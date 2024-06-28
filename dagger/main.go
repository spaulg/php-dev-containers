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
    Version      string // The major.minor.patch formatted version to build
    ShortVersion string // The major.minor formatted version to build
    MajorVersion string // The major version to build
    MinorVersion string // The minor version to build
    PatchVersion string // Patch version to build
    Suffix       string // Suffix to include in package name
    //	Architectures []string // Array of architectures to build
    PackageName  string // Package name to build
    Distribution string
    BuildNumber  int

    SourceDirectory *Directory

    BuildContainerImage    string
    BuildDirectoryName     string
    BuildDirectoryPath     string
    BuildDirectoryRootPath string

    TargetBaseContainerImage            string
    TargetBuildContainerImageRepository string
    TargetBuildContainerImageTag        string
    TargetBuildContainerPlatform        Platform
    TargetArchitecture                  string

    NoCache         bool
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

func New() *PhpDevContainers {
    return &PhpDevContainers{
        TargetArchitecture:                  runtime.GOARCH,
        Distribution:                        defaultDistribution,
        BuildNumber:                         defaultBuildNumber,
        BuildDirectoryRootPath:              packageDirectoryBase,
        TargetBuildContainerImageRepository: targetBuildContainerImageRepository,
    }
}

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

func (m *PhpDevContainers) WithSourceDirectory(sourceDirectory *Directory) *PhpDevContainers {
    m.SourceDirectory = sourceDirectory
    return m
}

func (m *PhpDevContainers) WithOutputDirectory(outputDirectory *Directory) *PhpDevContainers {
    m.OutputDirectory = outputDirectory
    return m
}

//func (m *PhpDevContainers) WithPhpArchitectures(architectures string) *PhpDevContainers {
//	m.Architectures = strings.Split(architectures, ",")
//	return m
//}

func (m *PhpDevContainers) BuildPhpImage(ctx context.Context) (*Container, error) {
    // todo: validate that version, suffix, architectures have all been set

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
