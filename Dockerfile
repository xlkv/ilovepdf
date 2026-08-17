# Build stage using Go 1.25+
FROM golang:1.25-bookworm AS builder

ENV GOTOOLCHAIN=auto
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o ilovepdf_bot ./cmd/bot

# Production runtime stage
FROM debian:bookworm-slim

# Install LibreOffice, Tesseract OCR, Python3, and certificates
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    libreoffice \
    libreoffice-script-provider-python \
    tesseract-ocr \
    tesseract-ocr-eng \
    tesseract-ocr-rus \
    python3 \
    poppler-utils \
    && rm -rf /var/lib/apt-get/lists/*

WORKDIR /app

# Copy built binary, webapp static files, and assets from builder
COPY --from=builder /app/ilovepdf_bot .
COPY --from=builder /app/webapp ./webapp
COPY --from=builder /app/assets ./assets

# Create temp directory
RUN mkdir -p /tmp/ilovepdf

EXPOSE 8088

CMD ["./ilovepdf_bot"]
