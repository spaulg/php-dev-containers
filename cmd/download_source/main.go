package main

import (
	"log"
	"main/internal/pkg/build"
)

func main() {
	// Parse command arguments to capture build information
	buildParameters := build.ParseArguments()

	// Download source archive
	sourceArchiveFileName, err := build.DownloadSourceArchive(buildParameters)

	if err != nil {
		log.Fatal(err)
	}

	log.Println("Source archive download to " + sourceArchiveFileName)
}
