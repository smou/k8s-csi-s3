FROM alpine:3.23
LABEL maintainers="Christopher Schütze <https://github.com/smou>"
LABEL description="csi-s3 slim image"

RUN apk add --no-cache fuse mailcap rclone
RUN apk add --no-cache -X http://dl-cdn.alpinelinux.org/alpine/edge/community s3fs-fuse

ADD https://github.com/yandex-cloud/geesefs/releases/latest/download/geesefs-linux-amd64 /usr/bin/geesefs
RUN chmod 755 /usr/bin/geesefs

COPY _output/s3driver /s3driver
ENTRYPOINT ["/s3driver"]
