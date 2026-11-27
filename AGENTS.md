# Working on ginkit

Conventions for anyone, human or agent, changing this template. The README
explains what the pieces are; this file is about how to change them.

## Flow

- `pnpm dev` runs both apps. The API listens on 3101 and the web app on 3100,
  proxying `/api` to the API. Nothing needs a `.env` file.
- `apps/api/openapi.yaml` is committed. Never edit it by hand: `go generate`
  in `apps/api` rewrites it, Air does so on every rebuild, and `pnpm check`
  fails when the committed copy is stale. The web client in
  `apps/web/src/lib/api` is generated from it and is not committed.
- Before committing run `pnpm fmt`, `pnpm lint`, `pnpm check`, and
  `pnpm build`.
- Images are built together by `pnpm docker:build`, which gives both the same
  version. The web app's `/readyz` fails when the API's version differs.

## Code

- Comments explain how to use or extend something, or warn about something
  that will bite. They do not describe what the code visibly does. When in
  doubt, leave the comment out.
- Formatting is oxfmt at its defaults and gofmt. Do not add style options.
- Linting is type-aware oxlint; correctness findings fail `pnpm lint` and the
  rest warn. Fix a finding in the code first. A rule goes to `off` in
  `oxlint.config.ts` only when a library's types leave no way to satisfy it,
  with a comment saying which.
- An environment variable is one file in `apps/api/internal/env`, read
  through `lookup`, plus a row in the README table and a line in
  `.env.example`. The web app declares its variables in `src/env.ts` with a
  valibot schema.
- JSON bodies go through `encoding/json/v2`, configured in `internal/app`. A
  struct that must keep properties it does not declare pairs
  `_ struct{} \`additionalProperties:"true"\``with`Extra map[string]any \`json:",embed" hidden:"true"\``, as `v1.Customer` does.
- API routes live in a subpackage per version under
  `apps/api/internal/handlers`; a new one is a file with a register function
  called from that version's `Register`.
- A client for another service is a package under `apps/api/internal/clients`,
  built once in `cmd/api/main.go` and set on `handlers.Dependencies`. The
  version that calls it takes an interface it declares, never the concrete
  client, so handlers stay testable with a fake.
- In the web app, `#lib` imports end in `.js`, SvelteKit types come from
  `@sveltejs/kit/hooks`, and everything is a devDependency because the
  build bundles it.
- The web app logs through `src/lib/server/log.ts`, which writes the same
  text and JSON shapes as the API's slog. Do not use `console` on the server.
- Server-side code in the web app calls the API through the client set up in
  `src/hooks.server.ts`, with the client's own fetch. Passing SvelteKit's
  `fetch` adds an `Origin` header that the API's CORS rejects.
- Keep the template general. Nothing application-specific belongs here: no
  database, no deployment files, no auth.

## Commits

- The message is a single subject line. No body, no trailers, no co-author
  lines.
- Do not push unless asked.
