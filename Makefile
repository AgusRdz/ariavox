.PHONY: build test coverage clean cross install uninstall doctor changelog \
        release release-patch release-minor release-major

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

# --- OS/Arch detection ---
UNAME_S := $(shell uname -s)
ifeq ($(findstring MINGW,$(UNAME_S)),MINGW)
  GOOS ?= windows
else ifeq ($(findstring MSYS,$(UNAME_S)),MSYS)
  GOOS ?= windows
else ifeq ($(findstring Darwin,$(UNAME_S)),Darwin)
  GOOS ?= darwin
else
  GOOS ?= linux
endif
GOARCH ?= $(if $(filter arm64 aarch64,$(shell uname -m)),arm64,amd64)
EXT    := $(if $(filter windows,$(GOOS)),.exe,)
BINARY := bin/ariavox$(EXT)

ifeq ($(GOOS),windows)
  INSTALL_DIR ?= $(LOCALAPPDATA)/Programs/ariavox
else
  INSTALL_DIR ?= $(HOME)/.local/bin
endif

# --- Build ---
build:
	docker compose run --rm dev go build -ldflags="$(LDFLAGS)" -o $(BINARY) ./cmd/ariavox

# --- Test ---
test:
	docker compose run --rm dev sh -c "CGO_ENABLED=1 go test ./... -v -race"

coverage:
	docker compose run --rm dev go test -coverprofile=coverage.out ./...
	docker compose run --rm dev go tool cover -func=coverage.out

# --- Local install (builds natively, no Docker) ---
install:
	@mkdir -p bin
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags='$(LDFLAGS)' -o $(BINARY) ./cmd/ariavox
	@mkdir -p "$(INSTALL_DIR)"
	cp $(BINARY) "$(INSTALL_DIR)/ariavox$(EXT)"
	@echo "installed ariavox $(VERSION) ($(GOOS)/$(GOARCH)) to $(INSTALL_DIR)/ariavox$(EXT)"
	@# Register in current shell session via the binary itself
	@"$(INSTALL_DIR)/ariavox$(EXT)" doctor > /dev/null 2>&1 || true

uninstall:
	@rm -f "$(INSTALL_DIR)/ariavox$(EXT)"
	@echo "removed ariavox from $(INSTALL_DIR)"

doctor:
	docker compose run --rm dev go run ./cmd/ariavox doctor

# --- Cross-compile ---
cross:
	docker compose run --rm dev sh -c "\
		CGO_ENABLED=0 GOOS=linux   GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/ariavox-linux-amd64   ./cmd/ariavox && \
		CGO_ENABLED=0 GOOS=linux   GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o bin/ariavox-linux-arm64   ./cmd/ariavox && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/ariavox-darwin-amd64  ./cmd/ariavox && \
		CGO_ENABLED=0 GOOS=darwin  GOARCH=arm64 go build -ldflags='$(LDFLAGS)' -o bin/ariavox-darwin-arm64  ./cmd/ariavox && \
		CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags='$(LDFLAGS)' -o bin/ariavox-windows-amd64.exe ./cmd/ariavox"

clean:
	rm -rf bin/ coverage.out

# --- Changelog (requires git-cliff) ---
.PHONY: _require-git-cliff
_require-git-cliff:
	@command -v git-cliff >/dev/null 2>&1 || { echo "git-cliff is required. See https://git-cliff.org/docs/installation"; exit 1; }

changelog: _require-git-cliff
	git-cliff --output CHANGELOG.md
	@echo "updated CHANGELOG.md"

# --- Release helpers ---
CURRENT_TAG := $(shell git describe --tags --abbrev=0 2>/dev/null || echo v0.0.0)
MAJOR := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f1)
MINOR := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f2)
PATCH := $(shell echo $(CURRENT_TAG) | sed 's/^v//' | cut -d. -f3)

release:
	@BUMP=patch; \
	if git log $$(git describe --tags --abbrev=0)..HEAD --format="%s" | grep -qE '^feat(\(.*\))?!:'; then BUMP=major; \
	elif git log $$(git describe --tags --abbrev=0)..HEAD --format="%B" | grep -q 'BREAKING CHANGE'; then BUMP=major; \
	elif git log $$(git describe --tags --abbrev=0)..HEAD --format="%s" | grep -qE '^feat'; then BUMP=minor; fi; \
	echo "detected: $$BUMP"; \
	$(MAKE) release-$$BUMP

release-patch: _require-git-cliff
	@NEXT=v$(MAJOR).$(MINOR).$(shell echo $$(($(PATCH)+1))); \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git-cliff --tag $$NEXT --output CHANGELOG.md && \
	git add CHANGELOG.md && \
	git commit -m "chore: update changelog for $$NEXT" && \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; git reset --soft HEAD~1; echo "push failed — tag and commit rolled back"; exit 1; }

release-minor: _require-git-cliff
	@NEXT=v$(MAJOR).$(shell echo $$(($(MINOR)+1))).0; \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git-cliff --tag $$NEXT --output CHANGELOG.md && \
	git add CHANGELOG.md && \
	git commit -m "chore: update changelog for $$NEXT" && \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; git reset --soft HEAD~1; echo "push failed — tag and commit rolled back"; exit 1; }

release-major: _require-git-cliff
	@NEXT=v$(shell echo $$(($(MAJOR)+1))).0.0; \
	echo "$(CURRENT_TAG) -> $$NEXT"; \
	git-cliff --tag $$NEXT --output CHANGELOG.md && \
	git add CHANGELOG.md && \
	git commit -m "chore: update changelog for $$NEXT" && \
	git tag $$NEXT && \
	{ git push origin HEAD $$NEXT && echo "released $$NEXT"; } || { git tag -d $$NEXT; git reset --soft HEAD~1; echo "push failed — tag and commit rolled back"; exit 1; }
