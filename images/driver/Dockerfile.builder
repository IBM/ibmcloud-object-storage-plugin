FROM golang:1.11.5

# Default values
ARG git_commit_id=unknown
ARG build_date=unknown

WORKDIR /go/src/github.com/IBM/ibmcloud-object-storage-plugin
ADD . /go/src/github.com/IBM/ibmcloud-object-storage-plugin
RUN set -ex; cd /go/src/github.com/IBM/ibmcloud-object-storage-plugin/ && CGO_ENABLED=0 go install -v -ldflags "-X main.Version=${git_commit_id} -X main.Build=${build_date}" github.com/IBM/ibmcloud-object-storage-plugin/cmd/driver
CMD ["/bin/bash"]
