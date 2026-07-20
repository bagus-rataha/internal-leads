# syntax=docker/dockerfile:1

# ---- Frontend build ------------------------------------------------------
FROM node:22-alpine AS frontend
WORKDIR /src
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# ---- Backend build --------------------------------------------------------
FROM golang:1.25-alpine AS backend
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
COPY --from=frontend /src/dist ./internal/static/dist
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /out/seed ./cmd/seed
RUN CGO_ENABLED=0 go build -tags 'postgres' -o /out/migrate github.com/golang-migrate/migrate/v4/cmd/migrate

# ---- Runtime ---------------------------------------------------------
# Deliberately alpine, not scratch/distroless: a shell is required so
# Portainer's container Console can exec into the running app to run
# migrate/seed by hand, per this project's manual-deploy workflow.
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata \
    && addgroup -S app && adduser -S app -G app
WORKDIR /app
COPY --from=backend /out/api /out/seed /out/migrate ./
COPY --from=backend /src/migrations ./migrations
COPY --from=backend /src/seeds ./seeds
COPY docker/migrate-up.sh ./migrate-up.sh
RUN chmod +x ./migrate-up.sh && chown -R app:app /app
USER app
EXPOSE 3000
CMD ["./api"]
