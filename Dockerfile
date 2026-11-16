# DietNow release image: the Go API + the built React app in one container.
#
#   docker build --build-arg VERSION=$(git rev-parse HEAD) -t dietnow .
#
# The API serves the SPA from /app/web (STATIC_DIR) and JSON under /api, so the
# browser talks to a single origin and the session cookie needs no cross-site
# setup. Dependencies are installed strictly from the lockfiles (npm ci,
# go mod download + -mod=readonly), so a build is reproducible from a commit.

# --- 1. Frontend ---------------------------------------------------------------
FROM node:22-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
RUN npm run build

# --- 2. Backend ----------------------------------------------------------------
FROM golang:1.22-alpine AS api
WORKDIR /src/backend
ENV CGO_ENABLED=0 GOFLAGS=-mod=readonly
COPY backend/go.mod backend/go.sum ./
RUN go mod download && go mod verify
COPY backend/ ./
ARG VERSION=dev
RUN go build -trimpath -ldflags "-s -w -X main.version=${VERSION}" -o /out/api ./cmd/api \
 && go build -trimpath -ldflags "-s -w" -o /out/migrate ./cmd/migrate

# --- 3. Runtime ----------------------------------------------------------------
# distroless/static: no shell or package manager; runs as an unprivileged user.
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=api /out/api /out/migrate /app/
COPY --from=web /src/frontend/dist /app/web
ARG VERSION=dev
LABEL org.opencontainers.image.source="https://github.com/R3ZV/DietNow" \
      org.opencontainers.image.revision="${VERSION}" \
      org.opencontainers.image.title="dietnow"
ENV STATIC_DIR=/app/web \
    API_PORT=8080
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/api"]
