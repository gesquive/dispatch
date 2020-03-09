FROM gesquive/go-builder:latest AS builder

ENV APP=dispatch

# This requires that `make release-snapshot` be called first
COPY dist/ /dist/
RUN copy-release

RUN mkdir -p /etc/${APP}/targets
COPY docker/config.yml /etc/${APP}

# =============================================================================
FROM gesquive/docker-base:latest
LABEL maintainer="Gus Esquivel <gesquive@gmail.com>"

# Import from builder
COPY --from=builder /app/${APP} /app/
COPY --from=builder /etc/${APP} /etc/

WORKDIR /config
VOLUME /config
EXPOSE 2525/tcp

ENTRYPOINT ["run", "/app/dispatch"]
