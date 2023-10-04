FROM golang:1.20.8

# Default values
ARG git_commit_id=unknown
ARG build_date=unknown

WORKDIR /go/src/github.com/IBM/ibmcloud-object-storage-plugin
ADD . /go/src/github.com/IBM/ibmcloud-object-storage-plugin
RUN set -ex; cd /go/src/github.com/IBM/ibmcloud-object-storage-plugin/ && CGO_ENABLED=0 GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -mod=mod -v -ldflags "-X main.Version=${git_commit_id} -X main.Build=${build_date}" github.com/IBM/ibmcloud-object-storage-plugin/cmd/driver
CMD ["/bin/bash"]
