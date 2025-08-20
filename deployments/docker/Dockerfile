# Build stage for frontend
FROM node:18-alpine AS frontend-builder
WORKDIR /app/web
COPY web/package*.json ./
RUN npm ci --only=production
COPY web/ ./
RUN npm run build

# Build stage for backend
FROM golang:1.21-alpine AS backend-builder
RUN apk add --no-cache git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o dixitme cmd/server/main.go

# Final stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/dixitme .

# Copy frontend build
COPY --from=frontend-builder /app/web/build ./web/build

# Copy assets
COPY assets ./assets

# Create non-root user
RUN addgroup -g 1000 appgroup && \
    adduser -u 1000 -G appgroup -s /bin/sh -D appuser
USER appuser

EXPOSE 8080

CMD ["./dixitme"]
