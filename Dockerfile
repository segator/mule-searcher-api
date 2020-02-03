FROM alpine
RUN apk add ca-certificates && rm -rf /var/cache/apk/*
WORKDIR /app
COPY dist/kad-e2dk-searcher /app/kad-e2dk-searcher
ENTRYPOINT ["/app/kad-e2dk-searcher"]
CMD ["-h"]
