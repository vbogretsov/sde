FROM golang:1.24-alpine AS build
RUN apk add gcc musl-dev
WORKDIR /src
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -o app -ldflags="-s -w" ./cmd/app
RUN ls -la

FROM alpine:latest
COPY --from=build /src/app /usr/local/bin/app
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT [ "app" ]
