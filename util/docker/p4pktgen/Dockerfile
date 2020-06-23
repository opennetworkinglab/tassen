# Copyright 2020-present Open Networking Foundation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG B_MODEL_SHA

FROM p4lang/behavioral-model:latest@$B_MODEL_SHA

ENV DEPS \
    curl  \
    git \
    sudo \
    python-dev \
    build-essential

RUN apt update && apt install -y $DEPS

RUN git clone https://github.com/p4pktgen/p4pktgen.git /p4pktgen
RUN cd /p4pktgen && ./tools/install.sh
# Clearing APT cache after install.sh
RUN rm -rf /var/lib/apt/lists/*

COPY run_p4pktgen.sh /

WORKDIR /p4pktgen

ENTRYPOINT []
