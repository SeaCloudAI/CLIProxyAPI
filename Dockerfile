FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X 'main.Version=${VERSION}' -X 'main.Commit=${COMMIT}' -X 'main.BuildDate=${BUILD_DATE}'" -o ./CLIProxyAPI ./cmd/server/

FROM alpine:3.22.0

RUN apk add --no-cache tzdata

RUN mkdir -p /CLIProxyAPI /CLIProxyAPI/static /app/bin /app/config /root/.cli-proxy-api

COPY --from=builder ./app/CLIProxyAPI /CLIProxyAPI/CLIProxyAPI
COPY --from=builder ./app/CLIProxyAPI /app/bin/seacloud-cli-proxy-api

COPY config.example.yaml /CLIProxyAPI/config.example.yaml
COPY config.example.yaml /CLIProxyAPI/config.yaml
COPY config.example.yaml /app/bin/config.example.yaml
COPY config.example.yaml /app/bin/config.yaml
COPY config.example.yaml /app/config/config.example.yaml
COPY config.example.yaml /app/config/config.yaml
COPY static/management.html /CLIProxyAPI/static/management.html
COPY docker-auth/ /root/.cli-proxy-api/

ENV MANAGEMENT_STATIC_PATH=/CLIProxyAPI/static/management.html

WORKDIR /CLIProxyAPI

EXPOSE 8080

ENV TZ=Asia/Shanghai

RUN cp /usr/share/zoneinfo/${TZ} /etc/localtime && echo "${TZ}" > /etc/timezone

CMD ["./CLIProxyAPI"]
