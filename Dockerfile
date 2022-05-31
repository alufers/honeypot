FROM --platform=$BUILDPLATFORM golang:alpine AS builder
ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG HONEYPOT_VERSION

# Git is required for fetching the dependencies.
RUN apk update && apk add --no-cache git bash nodejs npm gcc   util-linux-dev musl-dev  && mkdir -p /build/honeypot

WORKDIR /build/honeypot

COPY go.mod go.mod
COPY go.sum go.sum
COPY honeypot-frontend/package.json honeypot-frontend/package.json
COPY honeypot-frontend/package-lock.json honeypot-frontend/package-lock.json


RUN go mod download -json && (cd honeypot-frontend && npm install)

COPY . .

ENV HONEYPOT_VERSION=$HONEYPOT_VERSION

RUN mkdir -p /app && (cd honeypot-frontend && npm run build) && GOOS=${TARGETPLATFORM%%/*} GOARCH=${TARGETPLATFORM##*/} \
    go build -ldflags='-s -w -extldflags="-static"' -o /app/honeypot

# RUN echo "Running on architecture: $(uname -m), BUILDPLATFORM=$BUILDPLATFORM, TARGETPLATFORM=$TARGETPLATFORM" && exit 1

FROM scratch AS bin-unix
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/honeypot /app/honeypot

LABEL org.opencontainers.image.description A docker image for the honeypot service.

ENTRYPOINT ["/app/honeypot"]
