.PHONY: build \
	build-prepare-dir \
	build-version \
	deps-update \
	unittests \
	gofmt \
	golint \
	govet \
	$(NULL)

TARGET_GOOS=linux
TARGET_GOARCH=amd64

CACHE_DIR="_cache"

# Export GO111MODULE=on to enable project to be built from within GOPATH/src
export GO111MODULE=on

all: build

clean:
	rm -rf _output

check: gofmt golint govet

build-prepare-dir:
	[ ! -d _output/bin ] && mkdir -p _output/bin || :

build-version:
	./hack/build/build.sh

build: build-prepare-dir build-version
	@echo "Building binary"
	env GOOS=$(TARGET_GOOS) GOARCH=$(TARGET_GOARCH) go build -ldflags="-s -w" -mod=vendor -o _output/bin/numainfo_exporter ./cmd/numainfo_exporter

deps-update:
	go mod tidy && \
	go mod vendor

unittests:
	GOFLAGS=-mod=vendor go test -v ./pkg/...

gofmt:
	@echo "Running gofmt"
	gofmt -s -l `find . -path ./vendor -prune -o -type f -name '*.go' -print`

golint:
	@echo "Running go lint"
	hack/lint.sh

govet:
	@echo "Running go vet"
	go vet ./...

