PACKAGE := github.com/wjam/aws_finder

.DEFAULT_GOAL := all
.PHONY := clean all fmt linux mac windows coverage release build

release_dir := bin/release/
go_files := $(shell find . -path ./vendor -prune -o -path '*/testdata' -prune -o -type f -name '*.go' -print)
commands := $(notdir $(shell find cmd/* -type d))
local_bins := $(addprefix bin/,$(commands))
mac_suffix := -darwin-amd64
mac_bins := $(addsuffix $(mac_suffix),$(addprefix $(release_dir),$(commands)))
linux_suffix := -linux-amd64
linux_bins := $(addsuffix $(linux_suffix),$(addprefix $(release_dir),$(commands)))
windows_suffix := -windows-amd64.exe
windows_bins := $(addsuffix $(windows_suffix),$(addprefix $(release_dir),$(commands)))

clean:
	# Removing all generated files...
	@rm -rf bin/ || true

bin/.vendor: go.mod go.sum
	# Downloading modules...
	@go mod download
	@mkdir -p bin/
	@touch bin/.vendor

bin/.generate: $(go_files) bin/.vendor
	@go generate ./...
	@touch bin/.generate

fmt: bin/.generate $(go_files)
	# Formatting files...
	@go run golang.org/x/tools/cmd/goimports -w $(go_files)

bin/.vet: bin/.generate $(go_files)
	go vet  ./...
	@touch bin/.vet

bin/.fmtcheck: bin/.generate $(go_files)
	# Checking format of Go files...
	@GOIMPORTS=$$(go run golang.org/x/tools/cmd/goimports -l $(go_files)) && \
	if [ "$$GOIMPORTS" != "" ]; then \
		go run golang.org/x/tools/cmd/goimports -d $(go_files); \
		exit 1; \
	fi
	@touch bin/.fmtcheck

bin/.coverage.out: bin/.generate $(go_files)
	@go test -cover -v -count=1 ./... -coverpkg=$(shell go list ${PACKAGE}/... | xargs | sed -e 's/ /,/g') -coverprofile bin/.coverage.tmp
	@mv bin/.coverage.tmp bin/.coverage.out

coverage: bin/.coverage.out
	@go tool cover -html=bin/.coverage.out

$(local_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 go build -o $@ $(PACKAGE)/cmd/$(basename $(@F))

$(mac_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(subst $(mac_suffix),,$(@F)))

$(linux_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(subst $(linux_suffix),,$(@F)))

$(windows_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o $@ $(PACKAGE)/cmd/$(basename $(subst $(windows_suffix),,$(@F)))

$(release_dir)sha256sums.txt: $(mac_bins) $(linux_bins) $(windows_bins)
	@cd $(release_dir) && shasum -a 256 $(subst $(release_dir),,$^) > sha256sums.txt

linux: $(linux_bins)
windows: $(windows_bins)
mac: $(mac_bins)
build: $(local_bins)
release: linux windows mac $(release_dir)sha256sums.txt

all: release build
