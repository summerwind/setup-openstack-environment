VERSION=1.0.0
COMMIT=$(shell git rev-parse --verify HEAD)

PACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /cmd/)
BUILD_FLAGS=-ldflags "-X main.VERSION=$(VERSION) -X main.COMMIT=$(COMMIT)"

.PHONY: all
all: build

.PHONY: build
build: vendor
	go build $(BUILD_FLAGS) .

.PHONY: test
test:
	go test -v $(PACKAGES)
	go vet $(PACKAGES)

.PHONY: clean
clean:
	rm -rf setup-openstack-environment

dist:
	mkdir -p dist
	
	GOARCH=amd64 GOOS=linux go build $(BUILD_FLAGS) .
	tar -czf release/setup-openstack-environment_linux_amd64.tar.gz setup-openstack-environment
	rm -rf setup-openstack-environment

vendor:
	glide install
