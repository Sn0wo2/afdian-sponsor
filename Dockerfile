FROM alpine:latest

ARG TARGETPLATFORM

RUN mkdir -p /opt/afdian-sponsor

WORKDIR /opt/afdian-sponsor

COPY $TARGETPLATFORM/afdian-sponsor /opt/afdian-sponsor/afdian-sponsor

RUN chmod +x /opt/afdian-sponsor/afdian-sponsor

ENTRYPOINT ["/opt/afdian-sponsor/afdian-sponsor"]
