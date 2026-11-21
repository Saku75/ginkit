# syntax=docker/dockerfile:1

FROM golang:1.27-alpine AS build
# npm registry the docs assets are fetched from.
ARG NPM_REGISTRY=https://registry.npmjs.org
WORKDIR /src
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go run ./internal/handlers/docs/fetch.go -o internal/handlers/docs/assets && \
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

FROM gcr.io/distroless/static-debian13:nonroot
COPY --from=build /out/api /api

ENV MODE=release \
    PORT=8080 \
    LOG_LEVEL=info \
    LOG_FORMAT=json

EXPOSE 8080
ENTRYPOINT ["/api"]
