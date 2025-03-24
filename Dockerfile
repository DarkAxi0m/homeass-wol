FROM golang:1.24 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0
RUN make build

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/homeass-wol .
ENV MQTT_BROKER=
ENV MQTT_CLIENT_ID=homeass-wol-client
ENV MQTT_USERNAME=
ENV MQTT_PASSWORD=
ENV CONFIG_FILE=servers.yaml
EXPOSE 8080
CMD ["./homeass-wol"]


	
