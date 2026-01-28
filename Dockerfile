FROM alpine:latest
LABEL maintainers="Christopher Schütze <https://github.com/smou>"
LABEL description="minio-csi-s3 slim image"

RUN apk add --no-cache ca-certificates

COPY _output/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
