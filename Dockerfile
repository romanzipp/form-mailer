FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o form-mailer .

# Final stage
FROM alpine:latest
LABEL org.opencontainers.image.title="form-mailer" \
  org.opencontainers.image.description="HTML form to email service" \
  org.opencontainers.image.vendor="Roman Zipp" \
  org.opencontainers.image.authors="Roman Zipp" \
  org.opencontainers.image.licenses="NOASSERTION" \
  org.opencontainers.image.url="https://codeberg.org/romanzipp/form-mailer" \
  org.opencontainers.image.source="https://codeberg.org/romanzipp/form-mailer" \
  org.opencontainers.image.documentation="https://codeberg.org/romanzipp/form-mailer"

RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/form-mailer .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./form-mailer"]
