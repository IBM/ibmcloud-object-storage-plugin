#! /bin/sh

if [ ! -d "/root/s3fs-fuse" ]; then
    echo "s3fs code path /root/s3fs-fuse does not exists."
    exit 1
fi

cd /root/s3fs-fuse/
./autogen.sh
./configure CPPFLAGS='-I/usr/local/opt/openssl/include'
make
