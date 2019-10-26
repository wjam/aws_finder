PACKAGE := github.com/wjam/aws_finder

.DEFAULT_GOAL := all
.PHONY := clean all fmt linux mac windows

go_files := $(shell find . -path ./vendor -prune -o -type f -name '*.go' -print)
commands := $(notdir $(shell find cmd/* -type d))
mac_bins := $(addprefix bin/darwin/,$(commands))
linux_bins := $(addprefix bin/linux/,$(commands))
windows_bins := $(addsuffix .exe,$(addprefix bin/windows/,$(commands)))

clean:
	# Removing all generated files...
	@rm -rf bin/ .fmtcheck .test .vendor || true

.vendor: go.mod go.sum
	# Downloading modules...
	@go mod download
	@touch .vendor

.generate: $(go_files) go.mod go.sum
	@go generate ./...
	@touch .generate

bin/goimports: .vendor
	# Building goimports...
	@go build -o bin/goimports golang.org/x/tools/cmd/goimports

fmt: bin/goimports .generate $(go_files)
	# Formatting files...
	@bin/goimports -w .

.fmtcheck: bin/goimports .generate $(go_files)
	# Checking format of Go files...
	@GOIMPORTS=$$(bin/goimports -l .) && \
	if [ "$$GOIMPORTS" != "" ]; then \
		bin/goimports -d .; \
		exit 1; \
	fi
	@touch .fmtcheck

.test: .generate $(go_files)
	@go test -v -count=1 ./...
	@touch .test

$(mac_bins): .fmtcheck .test $(go_files)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

$(linux_bins): .fmtcheck .test $(go_files)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

$(windows_bins): .fmtcheck .test $(go_files)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

linux: $(linux_bins)
windows: $(windows_bins)
mac: $(mac_bins)

all: linux windows mac
