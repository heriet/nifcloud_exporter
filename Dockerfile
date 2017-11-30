FROM alpine:latest

COPY nifcloud_exporter /bin/
COPY config.yml /etc/nifcloud_exporter/config.yml

RUN apk update && \
    apk add ca-certificates && \
    update-ca-certificates

EXPOSE 9042
ENTRYPOINT [ "/bin/nifcloud_exporter", "-config.file=/etc/nifcloud_exporter/config.yml"]
