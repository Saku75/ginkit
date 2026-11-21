# ginkit

A monorepo template for a Go API with a SvelteKit frontend. The API is Gin
with Huma, which generates an OpenAPI document from the handlers. The frontend
is SvelteKit 3 on the Node adapter with Tailwind, Inter, and a typed client
generated from that document by hey-api. Turborepo and pnpm tie the two
together, and oxlint and oxfmt cover linting and formatting.

## Getting started

Requires Go 1.27 or newer, Node 24 or newer, and pnpm 12.

```sh
pnpm install
pnpm dev
```

That starts both apps:

| What          | Where                                           |
| ------------- | ----------------------------------------------- |
| Web app       | <http://localhost:3100>                         |
| API reference | <http://localhost:3100/api>, proxied to the API |
| API, directly | <http://localhost:3101>                         |
| Probes        | <http://localhost:3101/healthz> and `/readyz`   |

No `.env` files are needed in development. Each app has a `.env.example`
listing the variables it reads, with the defaults already set to these values.

To use this as a template, rename the pnpm scope and the Go module:

```sh
grep -rl '@ginkit/' --include='*.json' --include='*.ts' --include='Dockerfile' . | grep -v node_modules | xargs sed -i 's|@ginkit/|@yourscope/|g'
(cd apps/api && go mod edit -module yourscope/api && grep -rl 'ginkit/api/' --include='*.go' . | xargs sed -i 's|ginkit/api/|yourscope/api/|g')
```

## How the apps connect

- **The browser talks to `/api` on the web app's own origin.** Vite strips the
  prefix and forwards to the API on 3101, the same as a reverse proxy does in
  production. There is no CORS to configure.
- **The API serves at `/`** and does not know about the prefix. The proxy
  sends `X-Forwarded-Prefix: /api`, which the docs page uses for its links and
  the OpenAPI document for its server URL, so the reference and its "Try it"
  work both at the API's own address and under `/api`.
- **Server-side rendering calls the API directly** at `API_URL`, declared and
  validated in `apps/web/src/env.ts`. A relative URL would never reach the
  proxy from inside SvelteKit's server.
- **The OpenAPI document is committed** at `apps/api/openapi.yaml` and
  exported from the `@ginkit/api` package. `go generate` rewrites it, and Air
  does that on every Go change. The web app's `vite.config.ts` generates the
  client into `src/lib/api` from it on startup and again whenever it changes,
  so editing a Go handler updates the TypeScript types a moment later.

## Commands

Run these from the root. Turborepo runs the matching script in each app, in
dependency order, with the generate step first where needed.

| Command          | Does                                                                    |
| ---------------- | ----------------------------------------------------------------------- |
| `pnpm dev`       | Air for the API and Vite for the web app, side by side                  |
| `pnpm build`     | Builds both apps into `apps/api/bin` and `apps/web/build`               |
| `pnpm check`     | `svelte-check` on the web app; the API checks `openapi.yaml` is current |
| `pnpm lint`      | oxlint on the web app and `go vet` on the API                           |
| `pnpm test`      | Runs each app's tests                                                   |
| `pnpm fmt`       | oxfmt across the repo, then gofmt in the API                            |
| `pnpm fmt:check` | The same, without writing                                               |
| `pnpm generate`  | OpenAPI document and docs assets for the API                            |

## Layout

```
apps/api/
  generate.go         go generate directives: fetch the docs assets, then write openapi.yaml
  cmd/api/            Entry point: logging, signals, the server
  cmd/openapi/        Writes openapi.yaml; run by go generate
  internal/app/       Builds the router and Huma API from a Config
  internal/env/       One file per environment variable
  internal/handlers/  Probes, the docs page, and one subpackage per API version
  internal/logger/    slog setup from LOG_LEVEL and LOG_FORMAT
  internal/server/    HTTP server timeouts and the shutdown sequence
apps/web/
  src/env.ts          Environment variables, validated with valibot
  src/routes/+error.svelte  Error page
  src/hooks.server.ts Points the server's API client at API_URL; security headers; preloading
  src/app.css         Tailwind entry and the Inter font family
  src/lib/api/        Generated client, not committed
  src/lib/api-config.ts  Base URL for the browser's client
packages/             Empty until something is shared between the apps
turbo.json            Task graph
oxlint.config.ts      Lint rules at their defaults. .svelte files are linted in their script blocks.
oxfmt.config.ts       Formatting at its defaults, with .svelte files through Prettier's Svelte plugin.
```

## The API

| Variable               | Default       | Description                                                                         |
| ---------------------- | ------------- | ----------------------------------------------------------------------------------- |
| `MODE`                 | `development` | `development`, `staging`, or `release`. Development runs Gin in debug mode.         |
| `PORT`                 | `3101`        | Port to listen on. The image sets 8080.                                             |
| `BASE_PATH`            | none          | URL prefix to mount under, for hosting under a subpath without a proxy.             |
| `LOG_LEVEL`            | `info`        | `debug`, `info`, `warn`, or `error`.                                                |
| `LOG_FORMAT`           | `text`        | `text` or `json`.                                                                   |
| `TRUSTED_PROXIES`      | none          | Comma-separated IPs or CIDRs whose `X-Forwarded-For` header is trusted.             |
| `CLIENT_IP_HEADER`     | none          | Header carrying the real client IP, e.g. `CF-Connecting-IP`. Trusted from any peer. |
| `CORS_ALLOWED_ORIGINS` | none          | Comma-separated origins allowed from browsers, or `*`. Empty disables CORS.         |
| `SHUTDOWN_DELAY`       | `15s`         | How long to keep serving after SIGTERM while `/readyz` fails. `0s` disables.        |

| Route               | Purpose                                                    |
| ------------------- | ---------------------------------------------------------- |
| `GET /healthz`      | Liveness probe, 200 while the process is up                |
| `GET /readyz`       | Readiness probe, 503 when a check fails or during shutdown |
| `GET /`             | API reference, served from embedded files                  |
| `GET /openapi.yaml` | OpenAPI document (also `.json`)                            |
| `GET /v1`           | Root of API version 1                                      |
| `GET /v1/example`   | Example route                                              |

A route: add a file in `internal/handlers/v1` and call its register function
from `v1.Register`. A new API version: a subpackage next to `v1`, mounted
from `handlers.Register`. A readiness check: pass a `handlers.Check` to
`app.New` in `cmd/api/main.go`. An environment variable: copy a file in
`internal/env`.

## The web app

Environment variables are declared in `src/env.ts` with a valibot schema each
and read from `$app/env/private` or `$app/env/public`. SvelteKit validates
them at startup. `API_URL` is the only one so far.

Call the API through the generated functions, one per operation and named
after its operation ID. Pass SvelteKit's `fetch` in load functions:

```ts
import { v1Example } from "#lib/api/sdk.gen.js";

export const load: PageServerLoad = async ({ fetch }) => {
  const { data } = await v1Example({ fetch });
  return { message: data?.message ?? null };
};
```

The root layout imports Inter from `@fontsource-variable/inter` and
`src/app.css`, which imports Tailwind and sets the font family. The server hook
preloads the Latin font file along with scripts and styles. Everything the app
needs at runtime is bundled, so keep packages in `devDependencies`; anything in
`dependencies` would have to be installed into the image separately.

## Deploying

Each app builds its own image. The API image builds from its own directory
and the web image from the repository root, since it needs the workspace:

```sh
docker build -t ginkit-api apps/api
docker build -f apps/web/Dockerfile -t ginkit-web .
```

Both containers listen on 8080. The API logs JSON in release mode; pass
`--build-arg NPM_REGISTRY=...` to fetch its docs assets from another npm
registry. The web container needs `API_URL` pointing at the API, for example
`http://api:8080`, and `ORIGIN` set to its public URL.

Put a reverse proxy in front. Route `/api/*` to the API with the prefix
stripped and `X-Forwarded-Prefix: /api` added, which Traefik's StripPrefix
middleware does on its own and Caddy does with `handle_path` plus `header_up`.
Or give the API its own domain, where it needs no proxy at all. Point
Kubernetes probes at the API container's `/healthz` and `/readyz`.

On SIGINT or SIGTERM the API fails `/readyz` at once, keeps serving for
`SHUTDOWN_DELAY`, then stops accepting connections and gives in-flight requests
up to 15 seconds. A second signal exits immediately. On Kubernetes, set
`terminationGracePeriodSeconds` above the sum of the delay and those 15 seconds.
