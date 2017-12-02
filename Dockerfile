FROM golang:latest as builder
LABEL maintainer "heriet <heriet@heriet.info>"

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

COPY ./ /go/src/github.com/heriet/nifcloud_exporter/
WORKDIR /go/src/github.com/heriet/nifcloud_exporter/

RUN go get -u github.com/golang/dep/cmd/dep
RUN make


FROM alpine:latest

COPY --from=builder /go/src/github.com/heriet/nifcloud_exporter/nifcloud_exporter /bin/

COPY config.yml /etc/nifcloud_exporter/config.yml

RUN apk update && \
    apk add ca-certificates && \
    update-ca-certificates

EXPOSE 9042
ENTRYPOINT [ "/bin/nifcloud_exporter", "-config.file=/etc/nifcloud_exporter/config.yml"]
