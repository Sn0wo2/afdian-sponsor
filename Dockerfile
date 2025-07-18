FROM alpine:latest

RUN mkdir -p /opt/afdian-sponsor
WORKDIR /opt/afdian-sponsor

COPY afdian-sponsor /opt/afdian-sponsor/afdian-sponsor

RUN chmod +x /opt/afdian-sponsor/afdian-sponsor

ENTRYPOINT ["/opt/afdian-sponsor/afdian-sponsor"]
