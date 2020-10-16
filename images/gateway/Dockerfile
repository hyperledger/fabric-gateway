# Copyright IBM Corp. All Rights Reserved.
#
# SPDX-License-Identifier: Apache-2.0

ARG GO_VER
ARG ALPINE_VER

FROM alpine:${ALPINE_VER} as gateway-base
RUN apk add --no-cache tzdata
# set up nsswitch.conf for Go's "netgo" implementation
# - https://github.com/golang/go/blob/go1.9.1/src/net/conf.go#L194-L275
# - docker run --rm debian:stretch grep '^hosts:' /etc/nsswitch.conf
RUN echo 'hosts: files dns' > /etc/nsswitch.conf

FROM golang:${GO_VER}-alpine${ALPINE_VER} as golang
RUN apk add --no-cache \
	bash \
	gcc \
	git \
	make \
	musl-dev
ADD . $GOPATH/src/github.com/hyperledger/fabric-gateway
WORKDIR $GOPATH/src/github.com/hyperledger/fabric-gateway

FROM golang as gateway
ARG GO_TAGS
RUN make build-go GO_TAGS=${GO_TAGS}

FROM gateway-base
ENV FABRIC_CFG_PATH /etc/hyperledger/fabric-gateway
VOLUME /etc/hyperledger/fabric-gateway
VOLUME /var/hyperledger
COPY --from=gateway /go/src/github.com/hyperledger/fabric-gateway/bin/gateway /usr/local/bin
# COPY --from=gateway /go/src/github.com/hyperledger/fabric/sampleconfig/msp ${FABRIC_CFG_PATH}/msp
COPY --from=gateway /go/src/github.com/hyperledger/fabric-gateway/sampleconfig/gateway.yaml ${FABRIC_CFG_PATH}/config
COPY --from=gateway /go/src/github.com/hyperledger/fabric-gateway/sampleconfig/gateway.yaml ${FABRIC_CFG_PATH}
EXPOSE 7053
CMD ["gateway"]
