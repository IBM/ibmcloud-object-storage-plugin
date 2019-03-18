# ******************************************************************************
# * Licensed Materials - Property of IBM
# * IBM Cloud Container Service, 5737-D43
# * (C) Copyright IBM Corp. 2017, 2018 All Rights Reserved.
# * US Government Users Restricted Rights - Use, duplication or
# * disclosure restricted by GSA ADP Schedule Contract with IBM Corp.
# ******************************************************************************

PROVISIONER_CONT_IMG=armada-master/ibmcloud-object-storage-plugin
DRIVER_CONT_IMG=armada-master/ibmcloud-object-storage-driver

# TAG := $(shell git describe --abbrev=0 --tags HEAD 2>/dev/null)
# COMMIT := $(shell git rev-parse HEAD)

GOPACKAGES=$(shell go list ./... | grep -v '/cmd\|/pkg\|/s3fs-mount-health\|/vendor\|/tests')
GOFILES=$(shell find . -type f -name '*.go' -not -path "./vendor/*")

GIT_COMMIT_SHA="$(shell git rev-parse HEAD 2>/dev/null)"
GIT_REMOTE_URL="$(shell git config --get remote.origin.url 2>/dev/null)"
BUILD_DATE="$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")"
#IMAGE_TAG="$(shell awk '{print $$2}' VERSION)"
IMAGE_TAG := latest
# Jenkins vars. Set to `unknown` if the variable is not yet defined
BUILD_ID?=unknown
BUILD_NUMBER?=unknown
TRAVIS_BUILD_NUMBER?=unknown
#VERSION := latest

.PHONY: all
all: deps fmt vet test

.PHONY: fmt
fmt:
	gofmt -l ${GOFILES}
	@if [ -n "$$(gofmt -l ${GOFILES})" ]; then echo 'Above Files needs gofmt fixes. Please run gofmt -l -w on your code.' && exit 1; fi

.PHONY: vet
vet:
	go vet ${GOPACKAGES}

.PHONY: test
test:
	mkdir /tmp/s3fs
	$(GOPATH)/bin/gotestcover -v -race -coverprofile=cover.out ${GOPACKAGES}
	go tool cover -html=cover.out -o=cover.html
	rm -rf /tmp/s3fs

.PHONY: deps
deps:
	echo "Installing dependencies ..."
	glide install --strip-vendor
	go get github.com/pierrre/gotestcover

.PHONY: ut-coverage
ut-coverage: deps fmt vet test

# Build provisioner and driver images
.PHONY: plugin-images-build
plugin-images-build: build-provisioner-image build-driver-image
	# Tag and push image to user registry
	docker tag $(PROVISIONER_CONT_IMG):$(IMAGE_TAG) $(IMAGE_REGISTRY)/$(USER_NAMESPACE)/$(PLUGIN_IMAGE):$(PLUGIN_BUILD)
	docker push $(IMAGE_REGISTRY)/$(USER_NAMESPACE)/$(PLUGIN_IMAGE):$(PLUGIN_BUILD)

	docker tag $(DRIVER_CONT_IMG):$(IMAGE_TAG) $(IMAGE_REGISTRY)/$(USER_NAMESPACE)/$(DRIVER_IMAGE):$(DRIVER_BUILD)
	docker push $(IMAGE_REGISTRY)/$(USER_NAMESPACE)/$(DRIVER_IMAGE):$(DRIVER_BUILD)

.PHONY: build-provisioner-image
build-provisioner-image: deps
	# Spin provisioner builder container
	docker build -t provisioner-builder:latest -f images/provisioner/Dockerfile.builder . && sleep 30
	docker run provisioner-builder:latest  /bin/true
	docker cp `docker ps -q -n=1`:/root/ca-certs.tar.gz ./
	docker cp `docker ps -q -n=1`:/root/provisioner.tar.gz ./

	# Build Provisioner container image
	docker build \
        --build-arg git_commit_id=${GIT_COMMIT_SHA} \
        --build-arg git_remote_url=${GIT_REMOTE_URL} \
        --build-arg build_date=${BUILD_DATE} \
        --build-arg jenkins_build_id=${BUILD_ID} \
        --build-arg jenkins_build_number=${BUILD_NUMBER} \
	--build-arg travis_build_number=${TRAVIS_BUILD_NUMBER} \
        -t $(PROVISIONER_CONT_IMG):$(IMAGE_TAG) -f ./images/provisioner/Dockerfile .

	# Cleanup
	rm -f provisioner.tar.gz
	rm -f ca-certs.tar.gz

# Build Driver container image
.PHONY: build-driver-image
build-driver-image: build-fuse-binary copy-driver-binaries
	rm -rf s3fs-fuse
	docker build --build-arg git_commit_id=${GIT_COMMIT_SHA} \
        --build-arg git_remote_url=${GIT_REMOTE_URL} \
        --build-arg build_date=${BUILD_DATE} \
        --build-arg jenkins_build_id=${BUILD_ID} \
        --build-arg jenkins_build_number=${BUILD_NUMBER} \
	--build-arg travis_build_number=${TRAVIS_BUILD_NUMBER} \
        -t $(DRIVER_CONT_IMG):$(IMAGE_TAG)  ./pkg/

# Build fuse binary
.PHONY: build-fuse-binary
build-fuse-binary:
	git clone https://github.com/s3fs-fuse/s3fs-fuse.git
	docker build -t fuse:ubuntu16 -f ./images/driver/Dockerfile.ubuntu16 . && sleep 30
	docker run fuse:ubuntu16  /bin/true
	docker cp `docker ps -q -n=1`:/s3fs-fuse/src/s3fs ./s3fs16

	docker build -t fuse:ubuntu18 -f ./images/driver/Dockerfile.ubuntu18 . && sleep 30
	docker run fuse:ubuntu18  /bin/true
	docker cp `docker ps -q -n=1`:/s3fs-fuse/src/s3fs ./s3fs18

# Copy all binaries to be pushed with driver image
.PHONY: copy-driver-binaries
copy-driver-binaries: build-driver-binary
	echo "---Driver Version---" > version.txt
	$(GOPATH)/bin/ibmc-s3fs version >> version.txt
	echo "---s3fs Fuse Version---" >> version.txt
	./s3fs16 --version | head -n 1 >> version.txt
	CC=$(which musl-gcc) go build -o systemutil --ldflags '-w -linkmode external -extldflags "-static"' pkg/systemutil.go
	cp version.txt $(GOPATH)/bin/ibmc-s3fs s3fs16 s3fs18 systemutil ./pkg/

.PHONY: build-driver-binary
build-driver-binary: deps
	# Spin driver builder container and generate driver binary
	docker build --build-arg git_commit_id=${GIT_COMMIT_SHA} --build-arg build_date=${BUILD_DATE} \
	-t driver-builder  -f images/driver/Dockerfile.builder . && sleep 30
	docker run driver-builder /bin/true
	docker cp `docker ps -q -n=1`:/go/bin/driver $(GOPATH)/bin/ibmc-s3fs
	chmod 755 $(GOPATH)/bin/ibmc-s3fs

# Execute e2e testcases
.PHONY: test-binary-build-e2e
test-binary-build-e2e:
	#Install dependencies
	apt-get -y install mercurial
	cd ./tests/e2e; glide install -v
	go test ./tests/e2e/basic -c -o $(E2E_TEST_BINARY)

.PHONY: test-integration
test-integration:
	go test `go list ./... | grep -v 'vendor\|e2e'`

.PHONY: clean
clean:
	rm -f armada-storage-s3fs-plugin
