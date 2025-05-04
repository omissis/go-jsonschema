FROM golang:1.23.8-alpine3.21 AS tools

COPY scripts/tools-golang.sh /tmp/tools-golang.sh

RUN /tmp/tools-golang.sh && rm /tmp/tools-golang.sh

RUN apk add --no-cache jq~=1.7 yq~=4.33 && \
    rm -rf /var/cache/apk/* /tmp/*

FROM scratch AS go-jsonschema

ENTRYPOINT ["/go-jsonschema"]

COPY go-jsonschema /
