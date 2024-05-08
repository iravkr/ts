# SPDX-FileCopyrightText: 2019-present Open Networking Foundation <info@opennetworking.org>
# SPDX-FileCopyrightText: 2019-present Rimedo Labs
#
# SPDX-License-Identifier: Apache-2.0

FROM onosproject/golang-build:v1.3.0 as build

ENV GO111MODULE=on
ENV XAPPNAME=cco-mon

COPY Makefile go.mod go.sum /go/src/github.com/cco-mon/
COPY cmd/ /go/src/github.com/cco-mon/cmd/
COPY vendor/ /go/src/github.com/cco-mon/vendor/
COPY pkg/ /go/src/github.com/cco-mon/pkg/
COPY build/build-tools /go/src/github.com/cco-mon/build/build-tools/

RUN cd /go/src/github.com/cco-mon && GOFLAGS=-mod=vendor make build


FROM alpine:3.10
RUN apk add libc6-compat
RUN mkdir -p /etc/onos/config
COPY cco-mon/files/configs/config.json /etc/onos/config/config.json

USER nobody

COPY --from=build /go/src/github.com//cco-mon/build/_output/cco-mon /usr/local/bin/cco-mon


ENTRYPOINT ["cco-mon"]
