FROM alpine:edge
RUN apk add ca-certificates \
    && apk add -X http://dl-cdn.alpinelinux.org/alpine/edge/testing amule \
    && rm -rf /var/cache/apk/*

WORKDIR /app
COPY dist/kad-e2dk-searcher /app/kad-e2dk-searcher
ENTRYPOINT ["/app/kad-e2dk-searcher"]
CMD ["-h"]
