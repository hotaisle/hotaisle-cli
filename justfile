# =============================================================================
# 📚 Documentation
# =============================================================================
# This justfile provides a comprehensive build system for Go projects of any size.
# It supports development, testing, building, and deployment workflows.
#
# Quick Start:
# 1. Install 'just': https://github.com/casey/just
# 2. Copy this justfile to your project root
# 3. Run `just init` to initialize the project
# 4. Run `just --list` to see available commands
#
# Configuration:
# The justfile can be configured in several ways (in order of precedence):
# 1. Command line: just GOOS=darwin build
# 2. Environment variables: export GOOS=darwin
# 3. .env file in project root
# 4. Default values in this justfile

# =============================================================================
# 🔄 Core Configuration
# =============================================================================

set dotenv-load # Enable .env file support for local configuration
set positional-arguments # Allow passing arguments to recipes
set unstable

# Use bash with strict error checking
set shell := ["bash", "-uc"]

# Common command aliases for convenience
alias t := test
alias b := build
alias r := run
alias help := default


# =============================================================================
# Variables
# =============================================================================

# Project Settings
# These can be overridden via environment variables or .env file
project_name := env("PROJECT_NAME", "hotaisle-cli")
organization := env("ORGANIZATION", "hotaisle")
description := "Hot Aisle CLI"
maintainer := "hello@hotaisle.ai"
main_app := "."

build_platforms := "linux/amd64/- linux/arm64/- linux/arm/7 darwin/amd64/- darwin/arm64/- windows/amd64/- windows/arm64/-"


# Feature flags
# Enable/disable various build features
enable_docker := env("ENABLE_DOCKER", "false")
enable_docs := env("ENABLE_DOCS", "true")

# Build configuration
# Tags for conditional compilation
build_tags := ""
extra_tags := ""
all_tags := build_tags + " " + extra_tags

# Test configuration
# Settings for test execution and coverage
test_timeout := "5m"
coverage_threshold := "80"
bench_time := "2s"

# Go settings
# Core Go environment variables and configuration
go := require("go")
export GOPATH := env("GOPATH", shell(go + ' env GOPATH'))
export GOOS := env("GOOS", shell(go + ' env GOOS'))
export GOARCH := env("GOARCH", shell(go + ' env GOARCH'))
export CGO_ENABLED := env("CGO_ENABLED", "0")
gobin := GOPATH + "/bin"
go_version := shell(go + ' version')

# Automatically detect version information from git, unless release automation provides it.
version := env("VERSION", shell("git describe --tags --match 'v*' 2>/dev/null || echo dev"))

git_commit := trim(`git rev-parse --short HEAD 2>/dev/null || echo "unknown"`)
git_branch_raw := trim(shell('git rev-parse --abbrev-ref HEAD 2>/dev/null || echo ""'))
git_branch := if git_branch_raw == "" { "unknown" } else { git_branch_raw }
build_time := `date -u '+%Y-%m-%d_%H:%M:%S'`
build_by := `whoami`

# Directories
# Project directory structure
root_dir := justfile_directory()
bin_dir := root_dir + "/bin"
dist_dir := root_dir + "/dist"
dist_pkg_dir := root_dir + "/dist-pkg"
docs_dir := root_dir + "/docs"

# Build flags
# Linker flags for embedding version information
golist := shell(go + ' list -m')

ld_flags := ('-s -w' +
	" -X '" + golist + '/cmd/cli.Version=' 		+ version + "'" +
	" -X '" + golist + '/cmd/cli.Commit=' 		+ git_commit + "'" +
	" -X '" + golist + '/cmd/cli.Branch=' 		+ git_branch + "'" +
	" -X '" + golist + '/cmd/cli.BuildBy=' 		+ build_by + "'" +
	" -X '" + golist + '/cmd/cli.BuildTime='	+ build_time + "'" +
	" -X '" + golist + '/cmd/cli.GoVersion=' 	+ go_version + "'")

# =============================================================================
# Default Recipe
# =============================================================================

# Show available recipes with their descriptions
@default:
	just --list

# =============================================================================
# Project Setup
# =============================================================================

# Install all required development tools and dependencies
deps:
	#!/usr/bin/env bash
	set -euo pipefail

	{{go}} mod download

	# Install tools concurrently
	install_tool() {
		{{go}} install "$1"
	}
	export -f install_tool

	tools=(
		"github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
		"mvdan.cc/gofumpt@latest"
		"golang.org/x/vuln/cmd/govulncheck@latest"
		"github.com/golang/mock/mockgen@latest"
		"github.com/goreleaser/nfpm/v2/cmd/nfpm@latest"
	)

	printf '%s\n' "${tools[@]}" | xargs -P 0 -I {} bash -c 'install_tool "$@"' _ {}

# Update all project dependencies to their latest versions
deps-update:
	{{go}} get -u ./...
	{{go}} mod tidy

# =============================================================================
# Build
# =============================================================================

# Build the project
build:
	just build-one "{{bin_dir}}" "{{project_name}}" {{GOOS}} {{GOARCH}} "-"

build-one out_dir out_filename goos goarch goarm:
	#!/usr/bin/env bash
	set -euo pipefail

	out_path="{{out_dir}}/{{out_filename}}"
	# Find any existing output file (prefer compressed for timestamp check)
	newest_output=""
	[ -e "${out_path}.tar.gz" ] && newest_output="${out_path}.tar.gz" || \
	[ -e "${out_path}.zip" ] && newest_output="${out_path}.zip" || \
	[ -e "${out_path}" ] && newest_output="${out_path}"

	# Check if rebuild is needed (only if output exists)
	if [ -n "$newest_output" ] && [ -z "$(find . -name "*.go" -type f -newer "$newest_output" -print -quit 2>/dev/null)" ]; then
		echo "🛑 No changes detected, skipping ${out_path}"
		exit 0
	fi

	# Build
	mkdir -p "{{out_dir}}"
	GOOS={{goos}} GOARCH={{goarch}} GOARM={{goarm}} CGO_ENABLED={{CGO_ENABLED}} \
		{{go}} build \
		-tags '{{all_tags}}' \
		-ldflags="{{ld_flags}}" \
		-o "${out_path}" \
		./{{main_app}}
	echo "✅ GOOS={{goos}} GOARCH={{goarch}} GOARM={{goarm}} CGO_ENABLED={{CGO_ENABLED}} ldflags={{ld_flags}} ${out_path}"

build-all:
	#!/usr/bin/env bash
	set -euo pipefail
	platforms="{{build_platforms}}"

	build_platform() {
		local platform="$1"
		IFS='/' read -r os arch arm <<< "$platform"
		ext=""; [ "$os" = "windows" ] && ext=".exe"

		# Map 'arm' to 'armhf' for output filename (Debian convention)
		local out_arch="$arch"
		[ "$arch" = "arm" ] && out_arch="armhf"

		out_filename="{{project_name}}-{{version}}-${os}-${out_arch}${ext}"
		just build-one "{{dist_dir}}" "$out_filename" "$os" "$arch" "$arm"
	}
	export -f build_platform

	# Build all platforms concurrently
	echo "$platforms" | xargs -P 0 -n 1 bash -c 'build_platform "$@"' _
	echo "✅ build-all complete"

dist: build-all
	#!/usr/bin/env bash
	set -euo pipefail
	cd {{dist_dir}}

	# Find binaries (exclude already compressed files)
	shopt -s nullglob
	files=()
	for file in *; do
		[[ "$file" == *.gz || "$file" == *.zip ]] && continue
		[[ -f "$file" ]] && files+=("$file")
	done

	compress_file() {
		local file="$1"
		local compressed cmd

		# Determine compression method
		if [[ "$file" == *.exe ]]; then
			compressed="$file.zip"
			cmd="zip -q -9 '$compressed' '$file'"
		else
			compressed="$file.tar.gz"
			cmd="tar -czf '$compressed' '$file'"
		fi

		# Skip if compressed exists and is newer
		if [ -e "$compressed" ] && [ "$compressed" -nt "$file" ]; then
			echo "⏭️  Skipping $file (up to date)"
			return 0
		fi

		# Compress
		eval "$cmd" && echo "✅ Compressed $compressed"
	}
	export -f compress_file

	# Run compression in parallel
	printf '%s\n' "${files[@]}" | xargs -P 0 -I {} bash -c 'compress_file "$@"' _ {}
	echo "✅ Compression complete"

# Create deb and rpm packages for Linux ARM and AMD64 architectures
nfpm: build-all
	#!/usr/bin/env bash
	set -euo pipefail

	ARCHS="arm arm64 amd64"
	PACKAGERS="deb rpm"

	build_package() {
		local arch="$1"
		local packager="$2"

		# Map 'arm' to 'armhf' for Debian packages
		local file_arch="$arch"
		if [ "$arch" = "arm" ]; then
			file_arch="armhf"
		fi

		local binary="{{dist_dir}}/hotaisle-cli-{{version}}-linux-${file_arch}"
		local package="{{dist_pkg_dir}}/hotaisle-cli-{{version}}-linux-${file_arch}.${packager}"

		if [ -e "$package" ] && [ "$package" -nt "$binary" ]; then
			echo "⏭️  Skipping $package (up to date)"
			return 0
		fi

		mkdir -p "{{dist_pkg_dir}}"
		ARCH="$file_arch" VERSION="{{version}}" nfpm package \
			--packager "$packager" \
			--config package/nfpm.yaml \
			--target "$package"
		echo "✅ Created $package"
	}
	export -f build_package

	for ARCH in $ARCHS; do
		for PACKAGER in $PACKAGERS; do
			echo "$ARCH $PACKAGER"
		done
	done | xargs -P 0 -n 2 bash -c 'build_package "$@"' _
	echo "✅ nfpm complete"

# Update brew-formula.rb with version and SHA256 checksums
brew-formula:
	#!/usr/bin/env bash
	set -euo pipefail

	DARWIN_ARM64="{{dist_dir}}/hotaisle-cli-${VERSION}-darwin-arm64.tar.gz"
	DARWIN_AMD64="{{dist_dir}}/hotaisle-cli-${VERSION}-darwin-amd64.tar.gz"

	ARM64_SHA=$(sha256sum "$DARWIN_ARM64" | awk '{print $1}')
	AMD64_SHA=$(sha256sum "$DARWIN_AMD64" | awk '{print $1}')

	sed -e "s/VERSION/${VERSION}/g" \
		-e "s/ARM64_SHA256/${ARM64_SHA}/g" \
		-e "s/AMD64_SHA256/${AMD64_SHA}/g" \
		package/brew-formula.rb

# CI/CD
ci:	deps vet lint dist nfpm
release: deps dist nfpm

# Run the application
run *args: build
	{{bin_dir}}/{{project_name}} {{args}}

clean:
	@rm -rf {{dist_dir}} {{dist_pkg_dir}} {{bin_dir}}

# =============================================================================
# Testing & Quality
# =============================================================================

# Run tests
test:
	CGO_ENABLED=1 {{go}} test -v -race -cover ./...

# Run tests with coverage
test-coverage:
	{{go}} test -v -race -coverprofile=coverage.out ./...
	{{go}} tool cover -html=coverage.out -o coverage.html

# Run benchmarks
bench:
	{{go}} test -bench=. -benchmem -run=^$ -benchtime={{bench_time}} ./...

# Format code
fmt:
	{{go}} fmt ./...
	{{gobin}}/gofumpt -l -w .

# Run linters
lint:
	{{gobin}}/golangci-lint run --fix

# Run security scan
security:
	{{gobin}}/govulncheck ./...

# Run go vet
vet:
	{{go}} vet ./...

# [Local only] test that the github workflow ci.yml will run
act:
	act

# Push Docker image
docker-push:
	docker push {{organization}}/{{project_name}}:{{version}}

# Run Docker container
docker-run:
	docker run --rm -it {{organization}}/{{project_name}}:{{version}}

# Generate documentation
docs:
	mkdir -p {{docs_dir}}
	{{go}} doc -all > {{docs_dir}}/API.md

# Show version information
version:
	@echo "Version:     {{version}}"
	@echo "Commit:      {{git_commit}}"
	@echo "Branch:      {{git_branch}}"
	@echo "Build by:    {{build_by}}"
	@echo "Build time:  {{build_time}}"
	@echo "Go version:  {{go_version}}"
