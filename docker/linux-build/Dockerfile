# syntax=docker/dockerfile:1
FROM debian:latest
ENV HTTP_PROXY ""
ENV HTTPS_PROXY ""
ENV CGO_ENABLED "0"
ENV QPEP_REPO "https://github.com/Project-Faster/qpep.git"
ENV QPEP_BRANCH "main"
ENV QPEP_REV "HEAD"
ENV GOROOT "/opt/go"
ENV BUILDKIT_PROGRESS "plain"

USER root

RUN sh -c 'mkdir /build'

RUN sh -c 'echo export GOROOT=/opt/go'
RUN sh -c 'echo export PATH=$GOROOT/bin:$PATH'

RUN sh -c 'apt-get update'
RUN sh -c 'apt-get install -y wget git dos2unix'

WORKDIR /opt
RUN sh -c 'wget https://go.dev/dl/go1.18.10.linux-amd64.tar.gz'
RUN sh -c 'tar -xvvf go1.18.10.linux-amd64.tar.gz'

WORKDIR /root
RUN printf '#!/bin/bash -xe \n\
echo [CLONE] \n\
cd /root \n\
git clone ${QPEP_REPO} || true     \n\
cd /root/qpep \n\
git switch ${QPEP_BRANCH} \n\
git reset --hard ${QPEP_REV} \n\
export PATH=$GOROOT/bin:$PATH \n\
echo [BUILD] \n\
go build -v -o qpep \n\
cp /root/qpep/qpep /build/qpep \n\
echo [DONE] \n\
' > /root/build-qpep.sh

RUN sh -c 'dos2unix /root/build-qpep.sh'
RUN sh -c 'chmod 777 /root/build-qpep.sh'
RUN sh -c 'chmod +x /root/build-qpep.sh'

RUN sh -c 'cat /root/build-qpep.sh'

WORKDIR /root/qpep
ENTRYPOINT ["/root/build-qpep.sh"]
