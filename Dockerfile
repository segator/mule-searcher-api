FROM scratch
WORKDIR /app
COPY dist/kad-e2dk-searcher /app/kad-e2dk-searcher
RUN chmod +x /app/kad-e2dk-searcher
ENTRYPOINT /app/kad-e2dk-searcher
