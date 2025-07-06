FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./
COPY cmd ./cmd
COPY docs ./docs
COPY internal ./internal
COPY scripts ./scripts

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o vdlg -ldflags="-s -w" cmd/main.go

FROM gcr.io/distroless/static

WORKDIR /app

COPY --from=builder /app/vdlg ./
COPY --from=builder /app/scripts ./

EXPOSE 8080

ENTRYPOINT [ "/app/vdlg" ]
