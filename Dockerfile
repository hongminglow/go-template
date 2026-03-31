# syntax=docker/dockerfile:1.7

FROM golang:1.25-alpine AS build

WORKDIR /src
COPY go.mod ./

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/app ./cmd/api

FROM alpine:3.21

RUN addgroup -S app && adduser -S app -G app && \
    apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=build /out/app /app/app

USER app
EXPOSE 8080
CMD ["/app/app"]
