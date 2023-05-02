FROM golang:1.20.4-alpine3.17 AS tools

COPY scripts/tools-golang.sh /tmp/tools-golang.sh

RUN /tmp/tools-golang.sh && rm /tmp/tools-golang.sh

RUN apk add jq~=1.6 yq~=4.30

FROM scratch as go-jsonschema

ENTRYPOINT ["/go-jsonschema"]

COPY go-jsonschema /
