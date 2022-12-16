FROM scratch

ENTRYPOINT ["/go-jsonschema"]

COPY go-jsonschema /
