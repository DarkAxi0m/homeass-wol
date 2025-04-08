FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM alpine:latest

# Install ca-certificates, Lua 5.4, ping (iputils), and wakeonlan (via busybox-extras)
RUN apk --no-cache --virtual add \
    ca-certificates \
    lua5.4 lua5.4-libs lua5.4-socket \
    iputils \
    busybox-extras \
    && ln -s /usr/bin/lua5.4 /usr/bin/lua \
    && rm -rf /var/cache/apk/* /usr/share/doc /usr/share/man /usr/share/locale

WORKDIR /app
COPY --from=builder /app/bin/homeass-wol .
COPY ./scripts /app/scripts

ENV MQTT_BROKER=
ENV MQTT_CLIENT_ID=homeass-wol-client
ENV MQTT_USERNAME=
ENV MQTT_PASSWORD=
ENV CONFIG_FILE=servers.yaml

EXPOSE 8080
CMD ["./homeass-wol"]

