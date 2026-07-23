FROM golang:1.26.0
ADD . /go/src/github.com/IBM/ibmcloud-object-storage-plugin
RUN set -ex; cd /go/src/github.com/IBM/ibmcloud-object-storage-plugin/ && \
    echo "Starting go install..." && \
    CGO_ENABLED=0 go install -mod=mod -v github.com/IBM/ibmcloud-object-storage-plugin/cmd/provisioner | tee /tmp/build.log && \
    echo "Done."
RUN set -ex; tar cvC / ./etc/ssl  | gzip -n > /root/ca-certs.tar.gz
RUN ls -al /go/bin/
RUN set -ex; tar cvC /go/ ./bin | gzip -9 > /root/provisioner.tar.gz
