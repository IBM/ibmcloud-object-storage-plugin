# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

IMAGE = ibmcloud-object-storage-plugin

VERSION :=
TAG := $(shell git describe --abbrev=0 --tags HEAD 2>/dev/null)
COMMIT := $(shell git rev-parse HEAD)
GOPACKAGES=$(shell go list ./... | grep -v /vendor/ | grep -v /cmd)
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

GIT_COMMIT_SHA="$(shell git rev-parse HEAD 2>/dev/null)"
GIT_REMOTE_URL="$(shell git config --get remote.origin.url 2>/dev/null)"
BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"

#ifeq ($(TAG),)
#    VERSION := latest
#else
#    ifeq ($(COMMIT), $(shell git rev-list -n1 $(TAG)))
#        VERSION := $(TAG)
#    else
#        VERSION := $(TAG)-$(COMMIT)
#    endif
#endif
VERSION := latest
export GO111MODULE=on

.PHONY: all
all: deps fmt vet test

.PHONY: provisioner
provisioner: deps buildprovisioner

.PHONY: driver
driver: deps builddriver

.PHONY: deps
deps:
	echo "Installing dependencies ..."
	go mod download
	# go get github.com/coreos/go-systemd
	go install github.com/pierrre/gotestcover@latest

.PHONY: fmt
fmt:
	gofmt -l ${GOFILES}
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Above Files needs gofmt fixes. Please run gofmt -l -w on your code.' && exit 1; fi

.PHONY: vet
vet:
	go vet ${GOPACKAGES}

.PHONY: test
test:
	$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES}
	#go test -v -race -coverprofile=cover.out ${GOPACKAGES}

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o=cover.html

.PHONY: buildgo
buildgo:
	go build

.PHONY: buildprovisioner
buildprovisioner:
	# Build provisioner executable with explicit AMD64 platform
	docker build \
		--platform linux/amd64 \
		-t provisioner-builder \
		--pull \
		-f ./images/provisioner/Dockerfile.builder .
	
	# Create stopped container for reliable file copying
	docker create --name provisioner-builder-container provisioner-builder
	docker cp provisioner-builder-container:/root/ca-certs.tar.gz ./
	docker cp provisioner-builder-container:/root/provisioner.tar.gz ./
	docker rm provisioner-builder-container
	
	# Build final provisioner image for AMD64
	docker build \
		--platform linux/amd64 \
		--build-arg git_commit_id=${GIT_COMMIT_SHA} \
		--build-arg git_remote_url=${GIT_REMOTE_URL} \
		--build-arg build_date=${BUILD_DATE} \
		-t $(IMAGE):$(VERSION) \
		-f ./images/provisioner/Dockerfile .
	
	# Verify binary architecture
	@Echo "Verifying provisioner binary is AMD64..."
	@docker run --rm --entrypoint /bin/sh $(IMAGE):$(VERSION) -c \
		"file /usr/local/bin/provisioner | grep -q 'x86-64' && echo '✓ Correct AMD64 binary' || echo '✗ Wrong architecture!'"
	
	#Cleanup
	rm -f provisioner.tar.gz
	rm -f ca-certs.tar.gz

.PHONY: builddriver
builddriver:
	#Build and copy executables
	docker build --build-arg git_commit_id=${GIT_COMMIT_SHA} --build-arg build_date=${BUILD_DATE} -t driver-builder --pull -f images/driver/Dockerfile.builder .
	docker run driver-builder /bin/true
	docker cp `docker ps -q -n=1`:/go/bin/driver $(GOPATH)/bin/ibmc-s3fs
	chmod 755 $(GOPATH)/bin/ibmc-s3fs

.PHONY: push
push:
	docker push $(IMAGE):$(VERSION)

.PHONY: test-integration
test-integration:
	go test `go list ./... | grep -v 'vendor\|e2e'`

.PHONY: clean
clean:
	rm -f ibmcloud-object-storage-plugin
