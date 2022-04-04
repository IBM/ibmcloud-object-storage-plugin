FROM registry.access.redhat.com/ubi8/ubi-minimal:8.5-240

# Default values
ARG git_commit_id=unknown
ARG git_remote_url=unknown
ARG build_date=unknown
ARG jenkins_build_id=unknown
ARG jenkins_build_number=unknown
ARG travis_build_number=unknown

# Image Details
LABEL name="ibmcloud-object-storage-plugin"
LABEL vendor="IBM"
LABEL version="v1"
LABEL release="1.8.44"
LABEL summary="IBM COS Plugin plugin image"
LABEL description="Image to deploy ibmcloud-object-storage-plugin"
LABEL io.k8s.display-name="IBM COS Plugin"
LABEL io.k8s.description="Image to deploy ibmcloud-object-storage-plugin"
LABEL io.openshift.tags="1.8.44"
LABEL RUN="docker run icr.io/cpopen/ibmcloud-object-storage-plugin:1.8.44"
LABEL compliance.owner="ibm-armada-storage"


# Add Labels to image to show build details
LABEL git-commit-id=${git_commit_id}
LABEL git-remote-url=${git_remote_url}
LABEL build-date=${build_date}
LABEL jenkins-build-id=${jenkins_build_id}
LABEL jenkins-build-number=${jenkins_build_number}
LABEL travis_build_number=${travis_build_number}

#RUN mkdir /licenses
#ADD licenses /licenses
RUN microdnf update && microdnf install procps;
RUN microdnf install iputils
# Add the Provisioner executable
ADD ca-certs.tar.gz /
ADD provisioner.tar.gz /usr/local/
RUN chmod 755 /usr/local/bin/provisioner
USER 2121:2121
ENTRYPOINT ["/usr/local/bin/provisioner"]
