FROM golang:1.25.7-alpine3.23 AS tools

ARG TARGETPLATFORM

COPY scripts/tools-golang.sh /tmp/tools-golang.sh

RUN /tmp/tools-golang.sh && rm /tmp/tools-golang.sh

RUN apk add --no-cache jq~=1.8 yq~=4.53 && \
    rm -rf /var/cache/apk/* /tmp/*

FROM scratch AS go-jsonschema

ARG TARGETPLATFORM

ENTRYPOINT ["/go-jsonschema"]

COPY $TARGETPLATFORM/go-jsonschema /
