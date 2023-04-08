# syntax=docker/dockerfile:1
FROM debian:latest
ENV HTTP_PROXY ""
ENV HTTPS_PROXY ""
ENV CGO_ENABLED "0"
ENV QPEP_REPO "https://github.com/Project-Faster/qpep.git"
ENV QPEP_BRANCH "main"
ENV QPEP_REV "HEAD"
ENV GOROOT "/opt/go"

USER root

RUN sh -c 'mkdir /build'

RUN sh -c 'echo export GOROOT=/opt/go'
RUN sh -c 'echo export PATH=$GOROOT/bin:$PATH'

RUN sh -c 'apt-get update'
RUN sh -c 'apt-get install -y wget git'

WORKDIR /opt
RUN sh -c 'wget https://go.dev/dl/go1.18.10.linux-amd64.tar.gz'
RUN sh -c 'tar -xvvf go1.18.10.linux-amd64.tar.gz'

WORKDIR /root
RUN echo $'#!/bin/bash -xe \
    echo [CLONE] \
    git clone ${QPEP_REPO} || true     \
    cd /root/qpep \
    git switch ${QPEP_BRANCH} \
    git reset --hard ${QPEP_REV} \
    export PATH=$GOROOT/bin:$PATH \
    echo [BUILD] \
    go build -v -o qpep \
    cp /root/qpep/qpep /build/qpep \
    echo [DONE] \
    ' > /root/build-qpep.sh

RUN sh -c 'chmod 777 /root/build-qpep.sh'
RUN sh -c 'chmod +x /root/build-qpep.sh'

WORKDIR /root/qpep
ENTRYPOINT ["/root/build-qpep.sh"]
