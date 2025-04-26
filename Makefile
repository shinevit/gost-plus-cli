NAME=gost-tunnel
BINDIR=bin
VERSION=$(shell cat version/version.go | grep 'Version =' | sed 's/.*\"\(.*\)\".*/\1/g')
GOBUILD=CGO_ENABLED=1 go build --ldflags="-s -w" -v -a
GOFILES=*.go

# Get host architecture details
HOST_ARCH=$(shell uname -m)
HOST_OS=$(shell uname -s | tr '[:upper:]' '[:lower:]')

# Map host architecture to Go architecture
ifeq ($(HOST_ARCH),armv7l)
  GO_ARCH=arm
  GO_ARM=7
  PLATFORM=linux-armhf
else ifeq ($(HOST_ARCH),aarch64)
  GO_ARCH=arm64
  PLATFORM=linux-arm64
else ifeq ($(HOST_ARCH),x86_64)
  GO_ARCH=amd64
  PLATFORM=$(HOST_OS)-amd64
else
  GO_ARCH=$(HOST_ARCH)
  PLATFORM=$(HOST_OS)-$(HOST_ARCH)
endif

.DEFAULT_GOAL := default

# Create directory structure
$(BINDIR):
	@mkdir -p $(BINDIR)

# Default target is to build for the current platform
.PHONY: default
default: $(BINDIR) $(PLATFORM)

# Target for Raspberry Pi (armhf/ARMv7)
.PHONY: linux-armhf
linux-armhf: $(BINDIR)
	@echo "Building for Raspberry Pi (armhf/ARMv7)..."
	GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=1 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

# Standard targets for other platforms
.PHONY: linux-amd64
linux-amd64: $(BINDIR)
	@echo "Building for Linux (amd64)..."
	GOOS=linux GOARCH=amd64 CGO_ENABLED=1 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@"

.PHONY: linux-arm64
linux-arm64: $(BINDIR)
	@echo "Building for Linux (arm64)..."
	GOOS=linux GOARCH=arm64 CGO_ENABLED=1 go build --ldflags="-s -w" -v -o $(BINDIR)/$(NAME)-$(VERSION)-$@ $(GOFILES)
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
.PHONY: windows-amd64
windows-amd64: $(BINDIR)
	@echo "Building for Windows (amd64)..."
	GOOS=windows GOARCH=amd64 go-winres make --in winres/winres.json --out winres/rsrc
	@echo "Resource file generated"
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -ldflags="-s -w " -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@.exe"

.PHONY: windows-arm64
windows-arm64: $(BINDIR)
	@echo "Building for Windows (arm64)..."
	GOOS=windows GOARCH=arm64 go-winres make --in winres/winres.json --out winres/rsrc
	@echo "Resource file generated"
	GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -ldflags="-s -w" -o $(BINDIR)/$(NAME)-$(VERSION)-$@.exe $(GOFILES)
	@echo "Build complete: $(BINDIR)/$(NAME)-$(VERSION)-$@.exe"

.PHONY: android
android: $(BINDIR)
	@echo "Building for Android..."
	@if ! command -v gogio >/dev/null 2>&1; then \
		echo "Error: gogio not found. Please install with: go install gioui.org/cmd/gogio@latest"; \
		exit 1; \
	fi
	gogio -x -work -target android -minsdk 22 -version $(VERSION).8 -name GOST+ -signkey build/sign.keystore -signpass android -appid gost.plus -o $(BINDIR)/$(NAME)-$(VERSION).aab .
	gogio -x -work -target android -minsdk 22 -version $(VERSION).8 -name GOST+ -signkey build/sign.keystore -signpass android -appid gost.plus -o $(BINDIR)/$(NAME)-$(VERSION).apk .
	@echo "Build complete"

# Release packaging
.PHONY: package
package: $(BINDIR)
	@echo "Packaging $(PLATFORM)..."
	@if [[ "$(PLATFORM)" == *"windows"* ]]; then \
		zip -j $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).zip $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).exe; \
		echo "Package created: $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).zip"; \
	else \
		if [ -f "$(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)" ]; then \
			chmod +x $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM); \
			gzip -f -c $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM) > $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).gz; \
			echo "Package created: $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM).gz"; \
		else \
			echo "Error: Build file not found"; \
			exit 1; \
		fi; \
	fi

.PHONY: release
release: default package
	@rm -f $(BINDIR)/$(NAME)-$(VERSION)-$(PLATFORM)
	@echo "Release completed for $(PLATFORM)"

.PHONY: clean
clean:
	@echo "Cleaning build directory..."
	rm -f *.syso
	rm -rf $(BINDIR)
	@echo "Clean completed"

.PHONY: help
help:
	@echo "GOST+ Tunnel Build System"
	@echo "------------------------"
	@echo "Available targets:"
	@echo "  default       - Build for current platform ($(PLATFORM))"
	@echo "  linux-armhf   - Build for Raspberry Pi (ARMv7/armhf)"
	@echo "  linux-amd64   - Build for Linux AMD64"
	@echo "  linux-arm64   - Build for Linux ARM64"
	@echo "  darwin-amd64  - Build for macOS AMD64"
	@echo "  darwin-arm64  - Build for macOS ARM64"
	@echo "  windows-amd64 - Build for Windows AMD64"
	@echo "  windows-arm64 - Build for Windows ARM64"
	@echo "  android       - Build for Android"
	@echo "  package       - Package the build for current platform"
	@echo "  release       - Build and package for current platform"
	@echo "  clean         - Clean build artifacts"
