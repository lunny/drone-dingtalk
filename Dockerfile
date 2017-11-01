FROM alpine:3.5

RUN apk update && \
  apk add \
    ca-certificates && \
  rm -rf /var/cache/apk/*

ADD drone-dingtalk /bin/
ENTRYPOINT ["/bin/drone-dingtalk"]