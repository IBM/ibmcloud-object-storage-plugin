FROM busybox:1.26-glibc

# Default values
ARG git_commit_id=unknown
ARG git_remote_url=unknown
ARG build_date=unknown

# Add Labels to image to show build details
LABEL git-commit-id=${git_commit_id}
LABEL git-remote-url=${git_remote_url}
LABEL build-date=${build_date}

# Add the Provisioner executable
ADD ca-certs.tar.gz /
ADD provisioner.tar.gz /usr/local/
ENTRYPOINT ["/usr/local/bin/provisioner"]
