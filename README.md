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

| What          | Where                                                                                |
| ------------- | ------------------------------------------------------------------------------------ |
| Web app       | <http://localhost:3100>                                                              |
| API reference | <http://localhost:3100/api>, proxied to the API                                      |
| API, directly | <http://localhost:3101>                                                              |
| Probes        | `/healthz` and `/readyz` on both <http://localhost:3101> and <http://localhost:3100> |

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

| Command             | Does                                                                    |
| ------------------- | ----------------------------------------------------------------------- |
| `pnpm dev`          | Air for the API and Vite for the web app, side by side                  |
| `pnpm build`        | Builds both apps into `apps/api/bin` and `apps/web/build`               |
| `pnpm docker:build` | Builds both images with one version; see Versioning                     |
| `pnpm check`        | `svelte-check` on the web app; the API checks `openapi.yaml` is current |
| `pnpm lint`         | Type-aware oxlint on the web app and `go vet` on the API                |
| `pnpm test`         | Runs each app's tests                                                   |
| `pnpm fmt`          | oxfmt across the repo, then gofmt in the API                            |
| `pnpm fmt:check`    | The same, without writing                                               |
| `pnpm generate`     | OpenAPI document and docs assets for the API                            |

## Layout

```
apps/api/
  generate.go         go generate directives: fetch the docs assets, then write openapi.yaml
  cmd/api/            Entry point: logging, signals, the server
  cmd/openapi/        Writes openapi.yaml; run by go generate
  internal/app/       Builds the router and Huma API from a Config
  internal/clients/   One package per external service; rates is the example
  internal/env/       One file per environment variable
  internal/handlers/  Probes, the docs page, and one subpackage per API version
  internal/logger/    slog setup from LOG_LEVEL and LOG_FORMAT
  internal/server/    HTTP server timeouts and the shutdown sequence
  internal/version/   Build version, set by ldflags
apps/web/
  src/env.ts          Environment variables, validated with valibot
  src/routes/+error.svelte  Error page
  src/hooks.server.ts Points the server's API client at API_URL; security headers; preloading
  src/app.css         Tailwind entry and the Inter font family
  src/lib/api/        Generated client, not committed
  src/lib/api-config.ts  Base URL for the browser's client
packages/             Empty until something is shared between the apps
scripts/              build-images.sh builds both images with one version
turbo.json            Task graph
AGENTS.md             Conventions for contributors and coding agents
oxlint.config.ts      Type-aware oxlint: correctness rules fail the run, the suspicious, pedantic and perf
                      categories warn. .svelte files are linted in their script blocks.
oxfmt.config.ts       Formatting at its defaults, with .svelte files through Prettier's Svelte plugin.
```

## The API

| Variable               | Default                          | Description                                                                          |
| ---------------------- | -------------------------------- | ------------------------------------------------------------------------------------ |
| `MODE`                 | `development`                    | `development`, `staging`, or `release`. Development runs Gin in debug mode.          |
| `PORT`                 | `3101`                           | Port to listen on. The image sets 8080.                                              |
| `BASE_PATH`            | none                             | URL prefix to mount under, for hosting under a subpath without a proxy.              |
| `LOG_LEVEL`            | `info`                           | `debug`, `info`, `warn`, or `error`. `debug` also logs probe and version requests.   |
| `LOG_FORMAT`           | `text`                           | `text` or `json`.                                                                    |
| `TRUSTED_PROXIES`      | none                             | Comma-separated IPs or CIDRs whose `X-Forwarded-For` header is trusted.              |
| `CLIENT_IP_HEADER`     | none                             | Header carrying the real client IP, e.g. `CF-Connecting-IP`. Trusted from any peer.  |
| `CORS_ALLOWED_ORIGINS` | none                             | Comma-separated origins allowed from browsers, or `*`. Empty disables CORS.          |
| `SHUTDOWN_DELAY`       | `15s`                            | How long to keep serving after SIGTERM while `/readyz` fails. `0s` disables.         |
| `RATES_API_URL`        | `https://api.frankfurter.dev/v1` | Exchange-rate API behind the example price route. Any Frankfurter-compatible server. |

| Route                         | Purpose                                                                         |
| ----------------------------- | ------------------------------------------------------------------------------- |
| `GET /healthz`                | Liveness probe, 200 while the process is up                                     |
| `GET /readyz`                 | Readiness probe, 503 when a check fails or during shutdown                      |
| `GET /version`                | Build version; the web app's `/readyz` compares it with its own                 |
| `GET /`                       | API reference, served from embedded files                                       |
| `GET /openapi.yaml`           | OpenAPI document (also `.json`)                                                 |
| `GET /v1`                     | Root of API version 1                                                           |
| `GET /v1/example`             | Example route                                                                   |
| `GET /v1/products`            | Fixed catalog, filtered and paged by query parameters, count in `X-Total-Count` |
| `GET /v1/products/{id}`       | One product, or 404                                                             |
| `GET /v1/products/{id}/price` | The price in another currency, through the injected rates client                |
| `POST /v1/quotes`             | Prices a validated body; nothing is stored                                      |

A route: add a file in `internal/handlers/v1` and call its register function
from `v1.Register`. A new API version: a subpackage next to `v1`, mounted
from `handlers.Register`. A readiness check: pass a `handlers.Check` to
`app.New` in `cmd/api/main.go`, as the `rates` check shows; a check against
another service is right only when the API cannot work without it. An
environment variable: copy a file in `internal/env`.

A client for another service: a package under `internal/clients`, built once
in `cmd/api/main.go` and set on `handlers.Dependencies`. The version that
calls it declares the interface it needs, as `v1.Rates` does, and
`handlers.Register` passes the client in. The `rates` client and
`GET /v1/products/{id}/price` are the example of that path.

The products and quotes routes are examples to delete. `POST /v1/quotes`
shows how a body controls unknown properties: the request object is strict,
which Huma does by default; `lines` is a list of one known shape; `notes` is
a map that takes any object; and `customer` allows extra properties through
a `_` field tagged `additionalProperties:"true"`. Allowing them is not
enough to keep them: a struct drops properties it has no field for. So
`Customer` also has an `Extra map[string]any` tagged `json:",embed"`, which
`encoding/json/v2` fills with every unknown member and writes back on output,
and `hidden:"true"`, which keeps it out of the schema. The quote echoes the
customer back with the extras intact. The API reads and writes JSON with
`encoding/json/v2` for this, configured in `internal/app`.

## The web app

Environment variables are declared in `src/env.ts` with a valibot schema each
and read from `$app/env/private` or `$app/env/public`. SvelteKit validates
them at startup.

| Variable      | Default                 | Description                                                            |
| ------------- | ----------------------- | ---------------------------------------------------------------------- |
| `API_URL`     | `http://localhost:3101` | Where server-side rendering reaches the API.                           |
| `LOG_LEVEL`   | `info`                  | `debug`, `info`, `warn`, or `error`. `debug` also logs probe requests. |
| `LOG_FORMAT`  | `text`                  | `text` or `json`, the same two shapes as the API.                      |
| `APP_VERSION` | `dev`                   | Build version, inlined at build time. See Versioning.                  |

Every request is logged by `src/lib/server/hooks/logging.ts` and unexpected
errors by `handleError`, through `src/lib/server/log.ts`, which writes slog's
text and JSON formats so one pipeline reads both apps. Use it instead of
`console`.

`/healthz` answers 200 while the app is up. `/version` reports the build
version. `/readyz` runs the checks in `src/lib/server/checks.ts` and answers
in the API's format: 200 with each check's result, or 503 naming every check
that failed. The one check there calls the API's `/version` and fails while
the API is unreachable or built from a different version; add a check per
dependency the app cannot serve without.

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

## Versioning

The two apps are released in lockstep: one version string is built into both
images, and the web app refuses to become ready until the API it talks to
reports the same one, so a rolling deploy holds web traffic until the API has
rolled too. Locally everything reports `dev`.

```sh
pnpm docker:build          # version from git describe --tags --always --dirty
pnpm docker:build 1.4.0    # or an explicit one
```

That runs `scripts/build-images.sh`, which passes the same `VERSION` build
argument to both Dockerfiles. The API embeds it with `ldflags` into
`internal/version` and reports it from `/version` and as `info.version` of the
OpenAPI document it serves; the committed `openapi.yaml` is generated without
one and says `dev`. The web app inlines it as `APP_VERSION` at build time and
serves it from its own `/version`. Building one image on its own takes the same
argument: `docker build --build-arg VERSION=1.4.0 ...`.

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
Kubernetes probes at `/healthz` and `/readyz` on both containers.

On SIGINT or SIGTERM the API fails `/readyz` at once, keeps serving for
`SHUTDOWN_DELAY`, then stops accepting connections and gives in-flight requests
up to 15 seconds. A second signal exits immediately. On Kubernetes, set
`terminationGracePeriodSeconds` above the sum of the delay and those 15 seconds.
