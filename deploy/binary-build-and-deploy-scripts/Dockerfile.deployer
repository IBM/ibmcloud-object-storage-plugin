FROM nkkashyap/ibmutils:v002

ARG git_commit_id=unknown
ARG git_remote_url=unknown
ARG build_date=unknown

LABEL git-commit-id=${git_commit_id}
LABEL git-remote-url=${git_remote_url}
LABEL build-date=${build_date}

RUN apt-get update && apt-get install -y bash openssh-client
RUN mkdir -p /root/bin

ADD ./bin/ibmc-s3fs /root/bin
ADD ./bin/s3fs /root/bin

ADD install-driver.sh /root/bin
ADD install-dep.sh /root/bin

RUN chmod 775 /root/bin/install-driver.sh

CMD ["/root/bin/install-driver.sh"]
