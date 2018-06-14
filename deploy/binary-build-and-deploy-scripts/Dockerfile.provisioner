FROM busybox:1.26-glibc

ARG git_commit_id=unknown
ARG git_remote_url=unknown
ARG build_date=unknown

LABEL git-commit-id=${git_commit_id}
LABEL git-remote-url=${git_remote_url}
LABEL build-date=${build_date}

ADD ./bin/ca-certs.tar.gz /
ADD ./bin/provisioner /usr/local/bin/

RUN chmod 755 /usr/local/bin/provisioner

ENTRYPOINT ["/usr/local/bin/provisioner"]
