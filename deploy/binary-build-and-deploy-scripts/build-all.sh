#! /bin/bash
 
VERSION_TAG="v001"

echo -e "\nCloning repo..."
git clone https://github.com/IBM/ibmcloud-object-storage-plugin.git
git clone https://github.com/s3fs-fuse/s3fs-fuse.git

echo -e "\nSpinning builder image..."
docker build -t s3fs-plugin-builder:${VERSION_TAG} -f ./Dockerfile.build .

echo -e "\nCompiling s3fs fuse..."
docker run --name s3fsbuild-${VERSION_TAG} \
       -v `pwd`/s3fs-fuse:/root/s3fs-fuse s3fs-plugin-builder:${VERSION_TAG} /root/compile-s3fs.sh

echo -e "\nCompiling plugin..."
TARGET_PATH="/go/src/github.com/IBM/ibmcloud-object-storage-plugin"
docker run --name pluginbuild-${VERSION_TAG} \
       -v `pwd`/ibmcloud-object-storage-plugin:${TARGET_PATH} s3fs-plugin-builder:${VERSION_TAG} /root/compile-plugin.sh
if [[ $? -ne 0 ]]; then
   exit 1
fi

mkdir -p ./bin

cp s3fs-fuse/src/s3fs ./bin/
cp ibmcloud-object-storage-plugin/cmd/bin/*  ./bin/


BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
cd ./ibmcloud-object-storage-plugin 
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null)
GIT_REMOTE_URL=$(git config --get remote.origin.url 2>/dev/null)
cd ../

echo -e "\nSpinning provisioner image..."
docker build \
        --build-arg git_commit_id=${GIT_COMMIT} \
        --build-arg git_remote_url=${GIT_REMOTE_URL} \
        --build-arg build_date=${BUILD_DATE} \
        -t ibmcloud-object-storage-plugin:${VERSION_TAG} -f ./Dockerfile.provisioner .
if [[ $? -ne 0 ]]; then
   exit 1
fi

echo -e "\nSpinning deployer image..."
docker build \
        --build-arg git_commit_id=${GIT_COMMIT} \
        --build-arg git_remote_url=${GIT_REMOTE_URL} \
        --build-arg build_date=${BUILD_DATE} \
        -t ibmcloud-object-storage-deployer:${VERSION_TAG} -f ./Dockerfile.deployer .
if [[ $? -ne 0 ]]; then
   exit 1
fi

docker rm $(docker ps -q --filter status=exited)
