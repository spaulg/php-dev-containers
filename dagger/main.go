package main

import (
	"context"
	"dagger.io/dagger"
	"strings"
)

type PhpDevContainers struct {
	Version       string
	Suffix        string
	Architectures []string
	//PackageCache  *dagger.Directory
}

func (m *PhpDevContainers) WithPhpVersion(version string) *PhpDevContainers {
	m.Version = version
	return m
}

func (m *PhpDevContainers) WithPhpSuffix(suffix string) *PhpDevContainers {
	m.Suffix = suffix
	return m
}

func (m *PhpDevContainers) WithPhpArchitectures(architectures string) *PhpDevContainers {
	m.Architectures = strings.Split(architectures, ",")
	return m
}

//func (m *PhpDevContainers) WithPhpPackageCache(packageCache *dagger.Directory) *PhpDevContainers {
//	m.PackageCache = packageCache
//	return m
//}

func (m *PhpDevContainers) BuildPhpImage(ctx context.Context) *dagger.Container {
	dag, _ := dagger.Connect(ctx)
	container := dag.Container().
		From("alpine:latest").
		WithExec([]string{"sh", "-c", "echo 'version: " + m.Version + "' >> /debug"})

	container = container.
		WithExec([]string{"sh", "-c", "echo 'suffix: " + m.Suffix + "' >> /debug"})

	container = container.
		WithExec([]string{"sh", "-c", "echo 'architectures: " + strings.Join(m.Architectures, ", ") + "' >> /debug"})

	return container
}
