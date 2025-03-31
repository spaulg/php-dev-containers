package main

import (
	"context"
	"dagger/phpdevcontainers/internal/dagger"
	"encoding/json"
	"fmt"
	"github.com/dchest/uniuri"
	"io"
	"net/http"
	"strings"
)

func (m *PhpDevContainers) DownloadPhpSource(ctx context.Context) (*dagger.File, error) {
	sourceArchiveName := fmt.Sprintf("%s_%s.orig.tar.gz", m.PackageName, m.Version)

	sourceUrl, err := m.resolveDownloadUrl()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve source archive download url: %v", err)
	}

	container, err := dag.Container().
		From(m.BaseImage).
		Sync(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to download source archive: %v", err)
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

	container, err = container.
		WithEnvVariable("DEBIAN_FRONTEND", "noninteractive").
		WithExec([]string{"apt", "update", "-y"}).
		WithExec([]string{"apt", "upgrade", "-y"}).
		WithExec([]string{"apt", "install", "-y", "curl"}).
		WithExec([]string{"curl", "-fsSL", sourceUrl, "-o", sourceArchiveName}).
		Sync(ctx)

	if err != nil {
		return nil, fmt.Errorf("failed to download source archive: %v", err)
	}

	return container.File(sourceArchiveName), nil
}

func (m *PhpDevContainers) resolveDownloadUrl() (string, error) {
	// Fetch release metadata
	resp, err := http.Get("https://www.php.net/releases/index.php?json&version=" + m.Version)
	if err != nil {
		return "", fmt.Errorf("unable to request PHP release metadata: %v", err)
	}
	defer resp.Body.Close()

	// Read all response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("unable to read PHP release metadata response: %v", err)
	}

	// Unmarshall the response body json
	var data PhpVersion
	err = json.Unmarshal(bodyBytes, &data)
	if err != nil {
		return "", fmt.Errorf("unable to unmarshall PHP release metadata response: %v", err)
	}

	// Find the .tar.gz file
	for _, file := range data.Source {
		if strings.HasSuffix(file.Filename, ".tar.gz") {
			if data.Museum {
				return "https://museum.php.net/php" + m.MajorVersion + "/" + file.Filename, nil // museum
			} else {
				return "https://www.php.net/distributions/" + file.Filename, nil // latest
			}
		}
	}

	// Not found
	return "", fmt.Errorf("PHP Version %v not found", m.Version)
}
