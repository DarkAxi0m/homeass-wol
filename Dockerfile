FROM golang:1.22-alpine AS build

WORKDIR /app

COPY go.mod .
RUN go mod download

COPY . .

RUN go build -o mqtt-go-app main.go

FROM alpine:latest

WORKDIR /app
COPY --from=build /app/mqtt-go-app .

CMD ["./mqtt-go-app"]

