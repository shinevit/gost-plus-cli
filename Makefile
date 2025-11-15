NAME=gost-tunnel
BINDIR=bin
VERSION=$(shell grep 'Version =' version/version.go | cut -d'"' -f2)
GOBUILD=CGO_ENABLED=1 go build --ldflags="-s -w" -v -a
GOFILES=*.go

# Create bin directory
$(BINDIR):
	@mkdir -p $(BINDIR)

# Get host architecture details
HOST_ARCH=$(shell go env GOHOSTARCH)
HOST_OS=$(shell go env GOHOSTOS)

# Map host architecture to Go architecture
ifeq ($(HOST_ARCH),arm)
  GO_ARCH=arm
  GO_ARM=7
  PLATFORM=linux-armhf
else ifeq ($(HOST_ARCH),arm64)
  GO_ARCH=arm64
  PLATFORM=$(HOST_OS)-arm64
else ifeq ($(HOST_ARCH),amd64)
  GO_ARCH=amd64
  PLATFORM=$(HOST_OS)-amd64
else
  GO_ARCH=$(HOST_ARCH)
  PLATFORM=$(HOST_OS)-$(HOST_ARCH)
endif

WINDOWS_ARCH_LIST = \
        windows-386 \
        windows-amd64 \
        windows-arm64

OTHER_PLATFORM_LIST = \
        darwin-amd64 \
        darwin-arm64 \
        linux-armelv5 \
        linux-armelv6 \
        linux-armhf \
        linux-arm64 \
        linux-amd64 \

.DEFAULT_GOAL := default

# Default target is to build for the current platform
.PHONY: default
default: $(BINDIR) $(PLATFORM)

.PHONY: linux-armelv5
linux-armelv5: $(BINDIR)
		@echo "Building for Raspberry Pi (armel/ARMv5)..."
		GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
		@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

.PHONY: linux-armelv6
linux-armelv6: $(BINDIR)
		@echo "Building for Raspberry Pi (armel/ARMv6)..."
		GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
		@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

# Target for Raspberry Pi (ARMv7)
.PHONY: linux-armhf
linux-armhf: $(BINDIR)
	@echo "Building for Raspberry Pi (armhf/ARMv7)..."
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

# Standard targets for other platforms
.PHONY: linux-amd64
linux-amd64: $(BINDIR)
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

.PHONY: linux-arm64
linux-arm64: $(BINDIR)
	@echo "Building for Linux (aarch64 v8/9)..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

.PHONY: darwin-amd64
darwin-amd64: $(BINDIR)
	@echo "Building for macOS (amd64)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

.PHONY: darwin-arm64
darwin-arm64: $(BINDIR)
	@echo "Building for macOS (arm64)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

# Windows targets
.PHONY: windows-386
windows-386: $(BINDIR)
	@echo "Building for Windows (x86)..."
	GOOS=windows GOARCH=386 go-winres make --in winres/winres.json --out winres/rsrc
	@echo "Resource file generated"
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@.exe"

.PHONY: windows-amd64
windows-amd64: $(BINDIR)
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go-winres make --in winres/winres.json --out winres/rsrc
	@echo "Resource file generated"
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@.exe"

.PHONY: windows-arm64
windows-arm64: $(BINDIR)
	@echo "Building for Windows (arm64)..."
	GOOS=windows GOARCH=arm64 go-winres make --in winres/winres.json --out winres/rsrc
	@echo "Resource file generated"
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@.exe"

all: $(OTHER_PLATFORM_LIST) $(WINDOWS_ARCH_LIST)

# Build a release, compress and archive a binary
.PHONY: release
release: default
	@case "$(PLATFORM)" in \
		*windows*) \
		    if [ "$(PLATFORM)" = "windows-arm64" ]; then \
		        echo "Skipping UPX compression for Windows ARM64 (not supported)..."; \
		        cp $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).exe $(BINDIR)/$(NAME).exe; \
		    else \
		        echo "Compressing Windows .exe with UPX..."; \
		        upx -5 -o $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)-compressed.exe $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).exe; \
		        mv $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)-compressed.exe $(BINDIR)/$(NAME).exe; \
		    fi; \
			zip -9 -j $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).zip $(BINDIR)/$(NAME).exe; \
			echo "Archive created: $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).zip"; \
			rm -f $(BINDIR)/$(NAME).exe; \
			rm -f $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).exe; \
			;; \
		*darwin*) \
			echo "Compressing macOS binary with UPX (force-macos)..."; \
			upx --force-macos -5 -o $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)-compressed $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM); \
			mv $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)-compressed $(BINDIR)/$(NAME); \
			chmod +x $(BINDIR)/$(NAME); \
			GZIP=-9 tar -czf $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).tar.gz -C $(BINDIR) $(NAME); \
			echo "Archive created: $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).tar.gz"; \
			rm -f $(BINDIR)/$(NAME); \
			rm -f $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM); \
			;; \
		*) \
			if [ -f "$(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)" ]; then \
				echo "Compressing a binary with UPX..."; \
				upx -5 -o $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)-compressed $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM); \
				mv $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)-compressed $(BINDIR)/$(NAME); \
				chmod +x $(BINDIR)/$(NAME); \
				GZIP=-9 tar -czf $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).tar.gz -C $(BINDIR) $(NAME); \
				echo "Archive created: $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).tar.gz"; \
				rm -f $(BINDIR)/$(NAME); \
				rm -f $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM); \
			else \
				echo "Error: Build file not found"; \
				exit 1; \
			fi; \
			esac
	@echo "Release completed for $(PLATFORM)"

.PHONY: releases
releases: all
	@echo "Compressing binaries with UPX..."
	@# Compress Windows binaries
	@for platform in $(WINDOWS_ARCH_LIST); do \
		if [ -f "$(BINDIR)/$(NAME)-$(VERSION)-$${platform}.exe" ]; then \
			if [ "$${platform}" = "windows-arm64" ]; then \
				echo "Skipping UPX compression for Windows ARM64 (not supported)..."; \
			else \
				echo "Compressing Windows $${platform} binary..."; \
				upx -5 -o $(BINDIR)/$(NAME)-$(VERSION)-$${platform}-compressed.exe $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.exe >/dev/null 2>&1; \
				mv $(BINDIR)/$(NAME)-$(VERSION)-$${platform}-compressed.exe $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.exe; \
			fi; \
		fi; \
	done
	@# Compress other platform binaries
	@for platform in $(OTHER_PLATFORM_LIST); do \
		if [ -f "$(BINDIR)/$(NAME)-$(VERSION)-$${platform}" ]; then \
			case "$${platform}" in \
				*darwin*) \
					echo "Compressing macOS binary for $${platform} (force-macos)..."; \
					upx --force-macos -5 -o $(BINDIR)/$(NAME)-$(VERSION)-$${platform}-compressed $(BINDIR)/$(NAME)-$(VERSION)-$${platform} >/dev/null 2>&1; \
					;; \
				*) \
					echo "Compressing binary for $${platform}..."; \
					upx -5 -o $(BINDIR)/$(NAME)-$(VERSION)-$${platform}-compressed $(BINDIR)/$(NAME)-$(VERSION)-$${platform} >/dev/null 2>&1; \
					;; \
			esac; \
			mv $(BINDIR)/$(NAME)-$(VERSION)-$${platform}-compressed $(BINDIR)/$(NAME)-$(VERSION)-$${platform}; \
		fi; \
	done

	@echo "Packaging releases..."
	@# Package Windows releases
	@for platform in $(WINDOWS_ARCH_LIST); do \
		if [ -f "$(BINDIR)/$(NAME)-$(VERSION)-$${platform}.exe" ]; then \
			echo "Creating zip for $${platform}..."; \
			mv $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.exe $(BINDIR)/$(NAME).exe; \
			zip -9 -j $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.zip $(BINDIR)/$(NAME).exe; \
			rm -f $(BINDIR)/$(NAME).exe; \
			echo "Package created: $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.zip"; \
		fi; \
	done
	@# Package other platform releases
	@for platform in $(OTHER_PLATFORM_LIST); do \
		if [ -f "$(BINDIR)/$(NAME)-$(VERSION)-$${platform}" ]; then \
			echo "Creating tar.gz for $${platform}..."; \
			mv $(BINDIR)/$(NAME)-$(VERSION)-$${platform} $(BINDIR)/$(NAME); \
			chmod +x $(BINDIR)/$(NAME); \
			GZIP=-9 tar -czf $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.tar.gz -C $(BINDIR) $(NAME); \
			rm -f $(BINDIR)/$(NAME); \
			echo "Package created: $(BINDIR)/$(NAME)-$(VERSION)-$${platform}.tar.gz"; \
		fi; \
	done
	@echo "All releases compressed and packaged"

.SILENT:
.PHONY: clean
clean:
	echo "Cleaning build directory..."
	rm -f ./winres/*.syso > /dev/null 2>&1 || true
	rm -rf $(BINDIR) > /dev/null 2>&1 || true
	echo "Clean completed"

.PHONY: help
help:
	@echo "GOST+ Tunnel Build System"
	@echo "------------------------"
	@echo "Available targets:"
	@echo "  all           - Build for all platforms"
	@echo "  default       - Build for current platform ($(PLATFORM))"
	@echo "  linux-armelv5 - Build for ARMv5/armel"
	@echo "  linux-armelv6 - Build for ARMv6/armel"
	@echo "  linux-armhf   - Build for Raspberry Pi (ARMv7/armhf)"
	@echo "  linux-arm64   - Build for Raspberry Pi 3/4/5 ARM64 (ARMv8,ARMv9)"
	@echo "  linux-amd64   - Build for Linux AMD64"
	@echo "  darwin-amd64  - Build for macOS Intel-based (x86-64)"
	@echo "  darwin-arm64  - Build for macOS Apple Silicon (M1, M2, M3, etc.)"
	@echo "  windows-386   - Build for Windows x86"
	@echo "  windows-amd64 - Build for Windows AMD64"
	@echo "  windows-arm64 - Build for Windows ARM64"
	@echo "  package       - Package built binary for current platform"
	@echo "  release       - Build and package for current platform"
	@echo "  releases      - Build and package all release versions with appropriate archives"
	@echo "  clean         - Clean build artifacts"
