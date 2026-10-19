BUILD_DIR = builds
MODULE = github.com/soerenschneider/sc
BINARY_NAME = sc
CHECKSUM_FILE = $(BUILD_DIR)/checksum.sha256
DOCKER_PREFIX = ghcr.io/soerenschneider

# -s -w strip the symbol table and DWARF data, which stack traces do not need,
# and -trimpath removes local paths from the binary.
GO_BUILD = go build -trimpath -ldflags="-s -w -X '$(MODULE)/internal.BuildVersion=$(VERSION)' -X '$(MODULE)/internal.CommitHash=$(COMMIT_HASH)'"

generate:
	go generate  ./...

install: build
	cp sc ~/bin/

tests:
	go test ./... -race -covermode=atomic -coverprofile=coverage.out
	go tool cover -html=coverage.out -o=coverage.html
	go tool cover -func=coverage.out -o=coverage.out

clean:
	git diff --quiet || { echo 'Dirty work tree' ; false; }
	rm -rf ./$(BUILD_DIR)

build: version-info generate
	CGO_ENABLED=1 $(GO_BUILD) -o $(BINARY_NAME) .

release: clean version-info cross-build compress checksums

# Release binaries are published gzipped, which sc update decompresses with
# the standard library. -n leaves out name and timestamp for reproducibility.
compress:
	for f in $(BUILD_DIR)/$(BINARY_NAME)-*; do \
		case "$$f" in *.gz) ;; *) gzip -9 -n -f "$$f" ;; esac; \
	done

checksums: version-info
	{ echo "# version: $(VERSION)"; sha256sum $(BUILD_DIR)/sc-*; } > $(CHECKSUM_FILE)

# Signs keyless and writes $(CHECKSUM_FILE).sigstore.json. sc only trusts these
# signatures when made by the release workflow, so this only runs in CI.
sign:
	@if [ "$$CI" != "true" ]; then \
		echo "keyless signatures are only trusted from the release workflow"; false; \
	fi
	cosign sign-blob --yes --bundle $(CHECKSUM_FILE).sigstore.json $(CHECKSUM_FILE)

cross-build: build-linux build-darwin

# The darwin binary needs cgo for the keychain, so it can only be built on macOS.
build-darwin: version-info
	GOOS=darwin GOARCH=arm64 CGO_ENABLED=1      $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64   .

build-linux: version-info
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0       $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64    .
	GOOS=linux GOARCH=arm GOARM=6 CGO_ENABLED=0 $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-armv6    .
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0       $(GO_BUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-aarch64  .

docker-build:
	docker build -t "$(DOCKER_PREFIX)/sc-server" .

version-info:
	$(eval VERSION := $(shell git describe --tags --abbrev=0 || echo "dev"))
	$(eval COMMIT_HASH := $(shell git rev-parse HEAD))

fmt:
	find . -iname "*.go" -exec go fmt {} \; 

pre-commit-init:
	pre-commit install
	pre-commit install --hook-type commit-msg

pre-commit-update:
	pre-commit autoupdate

.PHONY: docs
docs:
	rm -rf docs/cli/*.md
	go run -tags=docgen main.go docs
