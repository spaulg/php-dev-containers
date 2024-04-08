package build

import (
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
	Version                   string
	ShortVersion              string
	MajorVersion              string
	MinorVersion              string
	PatchVersion              string
	Suffix                    string
	PackageName               string
	Distribution              string
	Architecture              string
	BuildNumber               int
	BuildContainerImage       string
	BuildDirectoryName        string
	BuildDirectoryPath        string
	TargetBaseContainerImage  string
	TargetBuildContainerImage string
	NoCache                   bool
}

const defaultDistribution = "bullseye"
const defaultBuildNumber = 1
const buildContainerImageRepository = "docker.io/spaulg/debuilder"
const targetBaseContainerImageRepository = "docker.io/debian"
const targetBuildContainerImageRepository = "docker.io/spaulg/php-ext-dev-containers"
const packagePrefix = "php"
const packageDirectoryBase = "/home/build/packages"

// ParseArguments parses the command line arguments and returns a BuildParameters struct of validated arguments
func ParseArguments() *BuildParameters {
	buildParameters := BuildParameters{
		Architecture: runtime.GOARCH,
		Distribution: defaultDistribution,
		BuildNumber:  defaultBuildNumber,
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

	flag.Func("architecture", "Build target architecture", func(s string) error {
		if len(strings.TrimSpace(s)) == 0 {
			return fmt.Errorf("--architecture argument cannot be empty")
		}

		buildParameters.Architecture = s
		return nil
	})

	flag.Func("distribution", "Debian build distribution", func(s string) error {
		if len(strings.TrimSpace(s)) == 0 {
			return fmt.Errorf("--distribution argument cannot be empty")
		}

		buildParameters.Distribution = s
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

	flag.Parse()

	// Version is a required field
	if buildParameters.Version == "" {
		log.Fatal("argument --version is required")
	}

	// Complete derived fields
	buildParameters.PackageName = packagePrefix + buildParameters.ShortVersion + buildParameters.Suffix
	buildParameters.BuildContainerImage = buildContainerImageRepository + ":" + buildParameters.Distribution
	buildParameters.BuildDirectoryName = buildParameters.PackageName + "_" + buildParameters.Version
	buildParameters.BuildDirectoryPath = packageDirectoryBase + "/" + buildParameters.BuildDirectoryName
	buildParameters.TargetBaseContainerImage = targetBaseContainerImageRepository + ":" + buildParameters.Distribution
	buildParameters.TargetBuildContainerImage = targetBuildContainerImageRepository + ":" + buildParameters.Distribution

	log.Println("Version: " + buildParameters.Version)
	log.Println("Short version: " + buildParameters.ShortVersion)
	log.Println("Suffix: " + buildParameters.Suffix)
	log.Println("PackageName: " + buildParameters.PackageName)
	log.Println("Distribution: " + buildParameters.Distribution)
	log.Println("Architecture: " + buildParameters.Architecture)
	log.Println("BuildNumber: " + strconv.Itoa(buildParameters.BuildNumber))
	log.Println("BuildContainerImage: " + buildParameters.BuildContainerImage)
	log.Println("TargetBaseContainerImage: " + buildParameters.TargetBaseContainerImage)
	log.Println("TargetBuildContainerImage: " + buildParameters.TargetBuildContainerImage)

	return &buildParameters
}
