PACKAGE := github.com/wjam/aws_finder

.DEFAULT_GOAL := all
.PHONY := clean all fmt linux mac windows coverage

go_files := $(shell find . -path ./vendor -prune -o -path '*/testdata' -prune -o -type f -name '*.go' -print)
commands := $(notdir $(shell find cmd/* -type d))
mac_bins := $(addprefix bin/darwin/,$(commands))
linux_bins := $(addprefix bin/linux/,$(commands))
windows_bins := $(addsuffix .exe,$(addprefix bin/windows/,$(commands)))

clean:
	# Removing all generated files...
	@rm -rf bin/ || true

bin/:
	@mkdir -p bin/

bin/.vendor: bin/ go.mod go.sum
	# Downloading modules...
	@go mod download
	@go mod tidy
	@touch bin/.vendor

bin/.generate: $(go_files) bin/.vendor go.mod go.sum
	@go generate ./...
	@touch bin/.generate

fmt: bin/.generate $(go_files)
	# Formatting files...
	@go run golang.org/x/tools/cmd/goimports -w $(go_files)

bin/.fmtcheck: bin/.generate $(go_files)
	# Checking format of Go files...
	@GOIMPORTS=$$(go run golang.org/x/tools/cmd/goimports -l $(go_files)) && \
	if [ "$$GOIMPORTS" != "" ]; then \
		go run golang.org/x/tools/cmd/goimports -d $(go_files); \
		exit 1; \
	fi
	@touch bin/.fmtcheck

bin/coverage.out: bin/.generate $(go_files)
	@go test -cover -v -count=1 ./... -coverprofile bin/coverage.out

coverage: bin/coverage.out
	@go tool cover -html=bin/coverage.out

$(mac_bins): bin/.fmtcheck bin/coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

$(linux_bins): bin/.fmtcheck bin/coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

$(windows_bins): bin/.fmtcheck bin/coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

linux: $(linux_bins)
windows: $(windows_bins)
mac: $(mac_bins)

all: linux windows mac
