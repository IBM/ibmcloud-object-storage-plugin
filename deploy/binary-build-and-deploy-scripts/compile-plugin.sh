#! /bin/sh

if [ ! -d "$GOPATH/src/github.com/IBM/ibmcloud-object-storage-plugin" ]; then
    echo "plugin code path $GOPATH/src/github.com/IBM/ibmcloud-object-storage-plugin does not exists."
    exit 1
fi
mkdir -p $GOPATH/bin
cd $GOPATH/src/github.com/IBM/ibmcloud-object-storage-plugin
mkdir -p ./cmd/bin

make
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null)
GIT_REMOTE_URL=$(git config --get remote.origin.url 2>/dev/null)

CGO_ENABLED=0 GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -mod=mod -v -ldflags "-X main.Version=${git_commit_id} -X main.Build=${build_date}" github.com/IBM/ibmcloud-object-storage-plugin/cmd/driver
CGO_ENABLED=0 GOOS=$(shell go env GOOS) GOARCH=$(shell go env GOARCH) go build -mod=mod -v github.com/IBM/ibmcloud-object-storage-plugin/cmd/provisioner

cd $GOPATH/src/github.com/IBM/ibmcloud-object-storage-plugin/cmd/bin
cp $GOPATH/bin/*  ./
cp ./driver ./ibmc-s3fs
tar cC / ./etc/ssl  | gzip -n > ./ca-certs.tar.gz
