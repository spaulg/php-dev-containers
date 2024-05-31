package build

import (
	"dagger.io/dagger"
	"flag"
	"fmt"
	"github.com/Masterminds/semver/v3"
	"log"
	"runtime"
	"strconv"
	"strings"
	"unicode"
)

type BuildParameters struct {
	Version      string
	ShortVersion string
	MajorVersion string
	MinorVersion string
	PatchVersion string
	Suffix       string
	PackageName  string
	Distribution string
	BuildNumber  int

	BuildContainerImage    string
	BuildDirectoryName     string
	BuildDirectoryPath     string
	BuildDirectoryRootPath string

	TargetBaseContainerImage            string
	TargetBuildContainerImageRepository string
	TargetBuildContainerImageTag        string
	TargetBuildContainerPlatform        dagger.Platform
	TargetArchitecture                  string

	NoCache             bool
	OutputDirectoryPath string
}

const defaultDistribution = "bullseye"
const defaultBuildNumber = 1
const buildContainerImageRepository = "docker.io/spaulg/debuilder"
const targetBaseContainerImageRepository = "docker.io/debian"
const targetBuildContainerImageRepository = "docker.io/spaulg/php-dev-containers"
const packagePrefix = "php"
const packageDirectoryBase = "/home/build/packages"
const outputDirectoryBase = "assets/packages"

// ParseArguments parses the command line arguments and returns a BuildParameters struct of validated arguments
func ParseArguments() *BuildParameters {
	buildParameters := BuildParameters{
		TargetArchitecture:                  runtime.GOARCH,
		Distribution:                        defaultDistribution,
		BuildNumber:                         defaultBuildNumber,
		BuildDirectoryRootPath:              packageDirectoryBase,
		TargetBuildContainerImageRepository: targetBuildContainerImageRepository,
	}

	flag.BoolFunc("no-cache", "No caching", func(s string) error {
		buildParameters.NoCache = true
		return nil
	})

	flag.Func("version", "PHP version", func(s string) error {
		version, err := semver.NewVersion(s)
		if err != nil {
			return fmt.Errorf("argument --version is not a valid semantic version: %v", err)
		}

		buildParameters.Version = fmt.Sprintf("%d.%d.%d", version.Major(), version.Minor(), version.Patch())
		buildParameters.ShortVersion = fmt.Sprintf("%d.%d", version.Major(), version.Minor())
		buildParameters.MajorVersion = fmt.Sprintf("%d", version.Major())
		buildParameters.MinorVersion = fmt.Sprintf("%d", version.Minor())
		buildParameters.PatchVersion = fmt.Sprintf("%d", version.Patch())

		return nil
	})

	flag.Func("suffix", "Package suffix", func(s string) error {
		if s != "" {
			for _, r := range s {
				if !unicode.IsDigit(r) && !unicode.IsLetter(r) {
					return fmt.Errorf("suffix must be alphanumeric")
				}
			}

			buildParameters.Suffix = "-" + s
		}

		return nil
	})

	flag.Func("build-number", "Build number", func(s string) error {
		var err error
		buildParameters.BuildNumber, err = strconv.Atoi(s)

		if err != nil {
			return fmt.Errorf("converting argument --build-number from string to int: %v", err)
		}

		if buildParameters.BuildNumber < 1 {
			return fmt.Errorf("--build-number argument cannot be less than 1")
		}

		return nil
	})

	flag.Func("distribution", "Debian build distribution", func(s string) error {
		if len(strings.TrimSpace(s)) == 0 {
			return fmt.Errorf("--distribution argument cannot be empty")
		}

		buildParameters.Distribution = s
		return nil
	})

	flag.Func("architecture", "Build target architecture", func(s string) error {
		if len(strings.TrimSpace(s)) == 0 {
			return fmt.Errorf("--architecture argument cannot be empty")
		}

		buildParameters.TargetArchitecture = s
		return nil
	})

	flag.Func("output-path", "Output path", func(s string) error {
		if s == "" {
			return fmt.Errorf("--output-path argument cannot be empty")
		}

		buildParameters.OutputDirectoryPath = s
		return nil
	})

	flag.Parse()

	// Version is a required field
	if buildParameters.Version == "" {
		log.Fatal("argument --version is required")
	}

	// Complete derived fields
	buildParameters.PackageName = packagePrefix + buildParameters.ShortVersion + buildParameters.Suffix
	buildParameters.BuildContainerImage = buildContainerImageRepository + ":" + buildParameters.Distribution
	buildParameters.BuildDirectoryName = buildParameters.PackageName + "_" + buildParameters.Version
	buildParameters.BuildDirectoryPath = buildParameters.BuildDirectoryRootPath + "/" + buildParameters.BuildDirectoryName
	buildParameters.TargetBaseContainerImage = targetBaseContainerImageRepository + ":" + buildParameters.Distribution

	buildParameters.TargetBuildContainerImageRepository = targetBuildContainerImageRepository
	buildParameters.TargetBuildContainerImageTag = buildParameters.PackageName + "-" + buildParameters.TargetArchitecture

	if buildParameters.OutputDirectoryPath == "" {
		buildParameters.OutputDirectoryPath = outputDirectoryBase + "/" + buildParameters.BuildDirectoryName +
			"-" + buildParameters.TargetArchitecture
	}

	var err error
	if buildParameters.TargetBuildContainerPlatform, err = mapContainerPlatform(buildParameters.TargetArchitecture); err != nil {
		log.Fatal(err)
	}

	log.Println("Version: " + buildParameters.Version)
	log.Println("Short version: " + buildParameters.ShortVersion)
	log.Println("Suffix: " + buildParameters.Suffix)
	log.Println("PackageName: " + buildParameters.PackageName)
	log.Println("BuildNumber: " + strconv.Itoa(buildParameters.BuildNumber))
	log.Println("Distribution: " + buildParameters.Distribution)

	log.Println("BuildContainerImage: " + buildParameters.BuildContainerImage)
	log.Println("BuildDirectoryRootPath: " + buildParameters.BuildDirectoryRootPath)
	log.Println("BuildDirectoryName: " + buildParameters.BuildDirectoryName)
	log.Println("BuildDirectoryPath: " + buildParameters.BuildDirectoryPath)

	log.Println("TargetArchitecture: " + buildParameters.TargetArchitecture)
	log.Println("TargetBaseContainerImage: " + buildParameters.TargetBaseContainerImage)
	log.Println("TargetBuildContainerImageRepository: " + buildParameters.TargetBuildContainerImageRepository)
	log.Println("TargetBuildContainerImageTag: " + buildParameters.TargetBuildContainerImageTag)

	log.Println("OutputDirectoryPath: " + buildParameters.OutputDirectoryPath)

	return &buildParameters
}

func mapContainerPlatform(targetPlatform string) (dagger.Platform, error) {
	switch targetPlatform {
	case "amd64":
		return "linux/amd64", nil
	case "arm64":
		return "linux/arm64", nil
	default:
		return "", fmt.Errorf("unsupport platform")
	}
}
