.DEFAULT_GOAL := all
.PHONY := clean all fmt linux mac windows coverage release build

release_dir := bin/release/
go_files := $(shell find . -path ./vendor -prune -o -path '*/testdata' -prune -o -type f -name '*.go' -print)
commands := $(shell go list -json ./... | jq -r '. | select(.Name == "main") | .Dir[(.Root | length) + 1:] | sub("^cmd/"; "")')
local_bins := $(addprefix bin/,$(commands))
mac_amd64_suffix := -darwin-amd64
mac_amd64_bins := $(addsuffix $(mac_amd64_suffix),$(addprefix $(release_dir),$(commands)))
mac_arm64_suffix := -darwin-arm64
mac_arm64_bins := $(addsuffix $(mac_arm64_suffix),$(addprefix $(release_dir),$(commands)))
linux_amd64_suffix := -linux-amd64
linux_amd64_bins := $(addsuffix $(linux_amd64_suffix),$(addprefix $(release_dir),$(commands)))
linux_arm64_suffix := -linux-arm64
linux_arm64_bins := $(addsuffix $(linux_arm64_suffix),$(addprefix $(release_dir),$(commands)))
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
	@go test -cover -v -count=1 ./... -coverpkg=./... -coverprofile bin/.coverage.tmp
	@mv bin/.coverage.tmp bin/.coverage.out

coverage: bin/.coverage.out
	@go tool cover -html=bin/.coverage.out

$(local_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 go build -trimpath -o $@ ./cmd/$(basename $(@F))

$(mac_amd64_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -o $@ ./cmd/$(basename $(subst $(mac_amd64_suffix),,$(@F)))

$(mac_arm64_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -o $@ ./cmd/$(basename $(subst $(mac_arm64_suffix),,$(@F)))

$(linux_amd64_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o $@ ./cmd/$(basename $(subst $(linux_amd64_suffix),,$(@F)))

$(linux_arm64_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -o $@ ./cmd/$(basename $(subst $(linux_arm64_suffix),,$(@F)))

$(windows_bins): bin/.fmtcheck bin/.vet bin/.coverage.out $(go_files)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -o $@ ./cmd/$(basename $(subst $(windows_suffix),,$(@F)))

$(release_dir)sha256sums.txt: $(mac_amd64_bins) $(mac_arm64_bins) $(linux_amd64_bins) $(linux_arm64_bins) $(windows_bins)
	@cd $(release_dir) && shasum -a 256 $(subst $(release_dir),,$^) > sha256sums.txt

linux: $(linux_amd64_bins) $(linux_arm64_bins)
windows: $(windows_bins)
mac: $(mac_amd64_bins) $(mac_arm64_bins)
build: $(local_bins)
release: linux windows mac $(release_dir)sha256sums.txt

all: release build
