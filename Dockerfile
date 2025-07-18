FROM alpine:latest

RUN addgroup -S sponsor-group && \
    adduser -S sponsor-user -G sponsor-group && \
    mkdir -p /opt/afdian-sponsor

WORKDIR /opt/afdian-sponsor

COPY --chown=sponsor-user:sponsor-group afdian-sponsor /opt/afdian-sponsor/afdian-sponsor

RUN chmod a-w /opt/afdian-sponsor/afdian-sponsor

USER sponsor-user

ENTRYPOINT ["./afdian-sponsor"]
