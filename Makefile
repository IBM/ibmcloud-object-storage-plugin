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

.PHONY: all
all: deps fmt vet test

.PHONY: provisioner
provisioner: deps buildprovisioner

.PHONY: driver
driver: deps builddriver

.PHONY: deps
deps:
	echo "Installing dependencies ..."
	glide install --strip-vendor
	go get github.com/pierrre/gotestcover

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

.PHONY: coverage
coverage:
	go tool cover -html=cover.out -o=cover.html

.PHONY: buildgo
buildgo:
	go build

.PHONY: buildprovisioner
buildprovisioner:
	#Build provisioner executable on target env
	docker build -t provisioner-builder --pull -f images/provisioner/Dockerfile.builder .
	docker run provisioner-builder /bin/true
	docker cp `docker ps -q -n=1`:/root/ca-certs.tar.gz ./
	docker cp `docker ps -q -n=1`:/root/provisioner.tar.gz ./

	#Make the final docker build having iscsilib and provisioner
	docker build \
        --build-arg git_commit_id=${GIT_COMMIT_SHA} \
        --build-arg git_remote_url=${GIT_REMOTE_URL} \
        --build-arg build_date=${BUILD_DATE} \
        -t $(IMAGE):$(VERSION) -f ./images/provisioner/Dockerfile .

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
