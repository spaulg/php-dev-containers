# PHP Dev Containers

This project is a [Dagger](https://dagger.io) module for building custom PHP packages and container images. 
It automates the process of downloading PHP source code, building Debian packages (`.deb`), and assembling 
multi-platform container images pushed to Docker Hub.

## Purpose

The primary goal is to provide a reproducible and scalable way to generate PHP containers based on 
Debian distributions for each major/minor version of PHP from 7.0+ upwards. The aim is for PHP module
development/debuging purposes.

Key goals:
 - all versions build for a single consistent Debian distribution (currently Bookworm)
 - scheduled rebuild every week on to refresh each version to ensure up to date container images
 - build both NTS and ZTS versions
 - build for both amd64 and arm64 architectures

## Prerequisites

- [Dagger CLI](https://docs.dagger.io/install)
- Docker or a compatible container runtime

## Dagger Command Examples

### Downloading source archives

Download PHP source archive:

```shell
dagger call \
  --distribution bookworm --version 8.5.1 \
  download-php-source \
  export --path work/php-8.5.1.tar.gz
```

Build packages, defaulting architecture to the current architecture:

```shell
dagger call \
  --distribution bookworm --version 8.5.1 \
  build-php-packages --source-archive work/php-8.5.1.tar.gz \
  export --path build/php-8.5.1
```

Build ZTS version

```shell
dagger call \
  --distribution bookworm --version 8.5.1+zts \
  build-php-packages --source-archive work/php-8.5.1.tar.gz \
  export --path build/php-8.5.1
```

Cross compile to multiple architectures

```shell
dagger call \
  --distribution bookworm --version 8.5.1 \
  build-php-packages --source-archive work/php-8.5.1.tar.gz --architectures="amd64,arm64" \
  export --path build/php-8.5.1
```

Cache bursting:

```shell
dagger call \
  --distribution bookworm --version 8.5.1 --burst-cache=true \
  build-php-packages --source-archive work/php-8.5.1.tar.gz \
  export --path build/php-8.5.1
```

Build container images:

```shell
dagger call \
  --distribution bookworm --version 8.5.1 \
  build-php-image --package-directory=build/php-8.5.1 --push=true
```

Build multiple architectures of container images (the current architecture will always be added if 
not included and will be built first):

```shell  
dagger call \
  --distribution bookworm --version 8.5.1 \
  build-php-image --package-directory=build/php-8.5.1 --push=true --architectures="amd64,arm64"
```
