# ═══════════════════════════════════════════════════════
#  Stage 1 — Build frontend
# ═══════════════════════════════════════════════════════
FROM node:22-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci

COPY frontend/ .
RUN npm run build

# ═══════════════════════════════════════════════════════
#  Stage 2 — Build backend
# ═══════════════════════════════════════════════════════
FROM golang:1.25-alpine AS backend-builder

WORKDIR /app/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o server cmd/server/main.go

# ═══════════════════════════════════════════════════════
#  Stage 3 — Runtime
# ═══════════════════════════════════════════════════════
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata curl

RUN adduser -D -h /app appuser
USER appuser
WORKDIR /app

COPY --from=backend-builder /app/backend/server .
COPY --from=frontend-builder /app/frontend/dist ./static

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --retries=3 \
  CMD curl -sf http://localhost:8080/api/health || exit 1

CMD ["./server"]
