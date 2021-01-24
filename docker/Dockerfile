FROM gesquive/go-builder:latest AS builder

ENV APP=dispatch

# This requires that `make release-snapshot` be called first
COPY dist/ /dist/
RUN copy-release
RUN chmod +x /app/dispatch

RUN mkdir -p /etc/dispatch/targets
COPY docker/config.yml /etc/dispatch

# =============================================================================
FROM gesquive/docker-base:busybox
LABEL maintainer="Gus Esquivel <gesquive@gmail.com>"

# Import from builder
COPY --from=builder /app/dispatch /app/
COPY --from=builder /etc/dispatch/ /etc/dispatch/

WORKDIR /config
VOLUME /config
EXPOSE 2525/tcp

ENTRYPOINT ["run"]
CMD ["/app/dispatch"]
