FROM alpine:3.23
LABEL maintainers="Christopher Schütze <https://github.com/smou>"
LABEL description="minio-csi-s3 slim image"

# Minimal runtime deps
RUN apk add --no-cache ca-certificates util-linux && \
    addgroup -S csi && \
    adduser -S -G csi -u 10001 csi

COPY _output/s3driver /usr/local/bin/s3driver
# Permissions
RUN chmod 0755 /usr/local/bin/s3driver
USER 10001:10001
ENTRYPOINT ["/usr/local/bin/s3driver"]
