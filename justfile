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

current_date := `date -u '+%Y%m%d-%H%M%S'`

# Version control
# Automatically detect version information from git
version := if `git rev-parse --git-dir 2>/dev/null; echo $?` == "0" {
	`git describe --tags --always --dirty 2>/dev/null || echo "dev"`
} else {
	"dev"
}

git_commit := trim(`git rev-parse --short HEAD 2>/dev/null || echo "unknown"`)
git_branch := trim(shell('git rev-parse --abbrev-ref HEAD 2>/dev/null || echo ""')) || "unknown"
build_time := `date -u '+%Y-%m-%d_%H:%M:%S'`
build_by := `whoami`

# Directories
# Project directory structure
root_dir := justfile_directory()
bin_dir := root_dir + "/bin"
dist_dir := root_dir + "/dist"
docs_dir := root_dir + "/docs"

# Build flags
# Linker flags for embedding version information
golist := shell(go + ' list -m')

ld_flags := '-s -w' + \
	" -X '" + golist + '/cmd/cli.Version=' 		+ version + "'" + \
	" -X '" + golist + '/cmd/cli.Commit=' 		+ git_commit + "'" + \
	" -X '" + golist + '/cmd/cli.Branch=' 		+ git_branch + "'" + \
	" -X '" + golist + '/cmd/cli.BuildBy=' 		+ build_by + "'" + \
	" -X '" + golist + '/cmd/cli.BuildTime='	+ build_time + "'" + \
	" -X '" + golist + '/cmd/cli.GoVersion=' 	+ go_version + "'"

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
	{{go}} mod download
	{{go}} install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	{{go}} install mvdan.cc/gofumpt@latest
	{{go}} install golang.org/x/vuln/cmd/govulncheck@latest
	{{go}} install github.com/golang/mock/mockgen@latest

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

build-one out_dir output goos goarch goarm:
	#!/usr/bin/env bash
	set -euo pipefail

	out_path="{{out_dir}}/{{output}}"
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
	GOOS={{goos}} GOARCH={{goarch}} CGO_ENABLED={{CGO_ENABLED}} GOARM={{goarm}} \
		{{go}} build \
		-tags '{{all_tags}}' \
		-ldflags="{{ld_flags}}" \
		-o "${out_path}" \
		./{{main_app}}
	echo "✅ GOOS={{goos}} GOARCH={{goarch}} CGO_ENABLED={{CGO_ENABLED}} GOARM={{goarm}} ${out_path}"

build-all:
	#!/usr/bin/env bash
	set -euo pipefail
	platforms="{{build_platforms}}"

	build_platform() {
		local platform="$1"
		IFS='/' read -r os arch arm <<< "$platform"
		ext=""; [ "$os" = "windows" ] && ext=".exe"
		output="{{project_name}}-{{version}}-${os}-${arch}${ext}"
		just build-one "{{dist_dir}}" "$output" "$os" "$arch" "$arm"
	}
	export -f build_platform

	# Build all platforms concurrently
	echo "$platforms" | xargs -P 0 -n 1 bash -c 'build_platform "$@"' _

dist:
	#!/usr/bin/env bash
	set -euo pipefail
	cd {{dist_dir}}
	files=$(find . -type f ! -name '*.gz' ! -name '*.tar' ! -name '*.zip' -exec basename {} \;)
	if [ -z "$files" ]; then
		echo "🛑 No files to compress"
		exit 0
	fi

	compress_file() {
		local file="$1"
		if [[ "$file" == *.exe ]]; then
			zip -q -9 "$file.zip" "$file" && rm "$file"
		else
			tar -cf "$file.tar" "$file" && gzip -f -9 "$file.tar" && rm "$file"
		fi
	}
	export -f compress_file

	# Run compression in parallel
	echo "$files" | xargs -P 0 -I {} bash -c 'compress_file "$@"' _ {}
	echo "✅ Compression complete"

# Run the application
run *args: build
	{{bin_dir}}/{{project_name}} {{args}}

ci:	deps build-all vet lint dist

clean:
	@rm -rf {{dist_dir}} {{bin_dir}}

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

# Update brew-formula.rb with version and SHA256 checksums
update-brew-formula:
	#!/usr/bin/env bash
	set -euo pipefail

	DARWIN_ARM64="{{dist_dir}}/hotaisle-cli-${VERSION}-darwin-arm64.tar.gz"
	DARWIN_AMD64="{{dist_dir}}/hotaisle-cli-${VERSION}-darwin-amd64.tar.gz"

	ARM64_SHA=$(sha256sum "$DARWIN_ARM64" | awk '{print $1}')
	AMD64_SHA=$(sha256sum "$DARWIN_AMD64" | awk '{print $1}')

	sed -e "s/VERSION/${VERSION}/g" \
		-e "s/ARM64_SHA256/${ARM64_SHA}/g" \
		-e "s/AMD64_SHA256/${AMD64_SHA}/g" \
		brew-formula.rb
