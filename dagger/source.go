package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func (m *PhpDevContainers) downloadSourceArchive() (string, error) {
	sourceArchiveName := fmt.Sprintf("%s_%s.orig.tar.gz", m.PackageName, m.Version)

	_, err := os.Stat("assets/source/" + sourceArchiveName)
	if err != nil {
		log.Println("Downloading source archive")

		if !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to detect source archive file status: %v", err)
		}

		sourceUrl, err := m.resolveDownloadUrl()

		if err != nil {
			return "", fmt.Errorf("failed to resolve source archive download url: %v", err)
		}

		out, err := os.Create("assets/source/" + sourceArchiveName)

		if err != nil {
			return "", fmt.Errorf("failed to create source archive file: %v", err)
		}
		defer out.Close()

		resp, err := http.Get(sourceUrl)

		if err != nil {
			return "", fmt.Errorf("failed to send HTTP request: %v", err)
		}
		defer resp.Body.Close()

		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to save source archive file to disk: %v", err)
		}
	}

	return sourceArchiveName, nil
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
