FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY internal ./internal

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vdlg -ldflags="-s -w" cmd/main.go

FROM alpine:latest AS ffmpeg-downloader

RUN apk add --no-cache curl tar xz

RUN curl -L https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz | tar -xJ && \
    mv ffmpeg-*-static/ffmpeg /ffmpeg

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /app/vdlg ./
COPY --from=ffmpeg-downloader /ffmpeg /usr/local/bin/ffmpeg

EXPOSE 8080

ENTRYPOINT [ "/app/vdlg" ]
