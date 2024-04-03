FROM golang:1.22.2-alpine3.18 AS tools

COPY scripts/tools-golang.sh /tmp/tools-golang.sh

RUN /tmp/tools-golang.sh && rm /tmp/tools-golang.sh

RUN apk add --no-cache jq~=1.6 yq~=4.33 && \
    rm -rf /var/cache/apk/* /tmp/*

FROM scratch as go-jsonschema

ENTRYPOINT ["/go-jsonschema"]

COPY go-jsonschema /
