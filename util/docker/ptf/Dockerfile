# Copyright 2019-present Open Networking Foundation
#
# SPDX-License-Identifier: Apache-2.0

# Docker image to run PTF-based data plane tests using stratum_bmv2

ARG MN_STRATUM_SHA

FROM bitnami/minideb:stretch as builder

ENV BUILD_DEPS \
    python-pip \
    python-setuptools \
    git
RUN install_packages $BUILD_DEPS

RUN mkdir -p /ouput

ENV PIP_DEPS \
    scapy==2.4.3 \
    git+https://github.com/p4lang/ptf.git \
    googleapis-common-protos==1.6.0 \
    ipaddress
RUN pip install --no-cache-dir --root /output $PIP_DEPS

FROM opennetworking/mn-stratum:latest@$MN_STRATUM_SHA as runtime

LABEL description="Docker image to run PTF-based data plane tests using stratum_bmv2"

ENV RUNTIME_DEPS \
    make
RUN install_packages $RUNTIME_DEPS

COPY --from=builder /output /

ENV DOCKER_RUN true

ENTRYPOINT []