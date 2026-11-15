# syntax=docker/dockerfile:1
# Build from the repository root: docker build -f apps/web/Dockerfile -t ginkit-web .

FROM node:24-alpine AS base
RUN npm install -g pnpm@12.21.0 turbo@2.11.11

FROM base AS pruner
WORKDIR /repo
COPY . .
RUN turbo prune @ginkit/web --docker

FROM base AS builder
WORKDIR /repo
COPY --from=pruner /repo/out/json/ .
RUN --mount=type=cache,target=/root/.local/share/pnpm/store pnpm install --frozen-lockfile
COPY --from=pruner /repo/out/full/ .
# Generates the client from the committed apps/api/openapi.yaml, so no Go toolchain is needed.
RUN pnpm --filter @ginkit/web build

FROM gcr.io/distroless/nodejs24-debian13:nonroot
WORKDIR /app
COPY --from=builder /repo/apps/web/build ./build
COPY --from=builder /repo/apps/web/package.json ./
ENV NODE_ENV=production \
    PORT=8080
EXPOSE 8080
CMD ["build"]
