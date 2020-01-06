FROM index.docker.io/gesquive/go-builder:latest AS builder

# This docker file relys on `make dist` being run first
ENV APP=dispatch
ARG TARGETARCH
ARG TARGETOS
ARG TARGETVARIANT

COPY dist/ /dist/
RUN copy-release

FROM scratch
LABEL maintainer="Gus Esquivel <gesquive@gmail.com>"

# Import from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd

COPY --from=builder /app/${APP} /app/

# Use an unprivileged user
USER runner

ENV DISPATCH_TARGET_DIR "/targets"
ENV DISPATCH_CONFIG "/config/config.yml"
ENV DISPATCH_LOG_FILE "stdout"
ENV DISPATCH_PORT 2525

VOLUME /targets
VOLUME /config

ENTRYPOINT ["/app/dispatch"]
