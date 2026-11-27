import { LOG_FORMAT, LOG_LEVEL } from "$app/env/private";

// Structured logging in the two formats the API's slog uses, so one pipeline
// reads both apps: text is `time=… level=INFO msg=… key=value`, json is one
// object per line with the same keys. Nested attributes become `a.b=value` in
// text and nested objects in json. Error values are logged as their message.
export type Attrs = Record<string, unknown>;
type Level = "debug" | "info" | "warn" | "error";

const order: Record<Level, number> = { debug: 0, info: 1, warn: 2, error: 3 };
const threshold = order[LOG_LEVEL];

export const log = {
  debug: logger("debug"),
  info: logger("info"),
  warn: logger("warn"),
  error: logger("error"),
  enabled: (level: Level) => order[level] >= threshold,
};

function logger(level: Level) {
  return (msg: string, attrs: Attrs = {}) => {
    if (order[level] < threshold) return;
    const time = new Date().toISOString();
    const name = level.toUpperCase();
    const clean = normalize(attrs);
    const line =
      LOG_FORMAT === "json"
        ? JSON.stringify({ time, level: name, msg, ...clean })
        : [`time=${time}`, `level=${name}`, `msg=${quote(msg)}`, ...flatten(clean)].join(" ");
    (order[level] >= order.warn ? process.stderr : process.stdout).write(line + "\n");
  };
}

const isRecord = (value: unknown): value is Attrs =>
  typeof value === "object" && value !== null && !Array.isArray(value);

function normalize(attrs: Attrs): Attrs {
  return Object.fromEntries(
    Object.entries(attrs)
      .filter(([, value]) => value !== undefined)
      .map(([key, value]): [string, unknown] => [key, normalizeValue(value)]),
  );
}

// Leaves what JSON can carry and turns the rest into strings, so both formats
// print the same thing and JSON.stringify never throws on a bigint.
function normalizeValue(value: unknown): unknown {
  if (value instanceof Error) return value.message;
  if (value instanceof Date) return value.toISOString();
  if (Array.isArray(value)) return value.map((item: unknown) => normalizeValue(item));
  if (isRecord(value)) return normalize(value);
  if (typeof value === "bigint" || typeof value === "symbol" || typeof value === "function") {
    return String(value);
  }
  return value;
}

function flatten(value: unknown, prefix = ""): string[] {
  if (isRecord(value)) {
    return Object.entries(value).flatMap(([key, item]) =>
      flatten(item, prefix ? `${prefix}.${key}` : key),
    );
  }
  return [`${prefix}=${quote(value)}`];
}

// slog quotes a value that is empty or holds whitespace, `=`, `"` or a control
// character, and prints nil for a missing one.
function quote(value: unknown): string {
  const text = plain(value);
  return text === "" || /[\s="\p{Cc}]/u.test(text) ? JSON.stringify(text) : text;
}

function plain(value: unknown): string {
  if (value === null || value === undefined) return "<nil>";
  if (typeof value === "string") return value;
  if (typeof value === "number" || typeof value === "boolean") return String(value);
  return JSON.stringify(value);
}
