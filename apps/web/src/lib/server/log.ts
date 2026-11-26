import { LOG_FORMAT, LOG_LEVEL } from "$app/env/private";

// Structured logging in the two formats the API's slog uses, so one pipeline
// reads both apps: text is `time=… level=INFO msg=… key=value`, json is one
// object per line with the same keys. Nested attributes become `a.b=value` in
// text and nested objects in json. Error values are logged as their message.
export type Attrs = Record<string, unknown>;
type Level = "DEBUG" | "INFO" | "WARN" | "ERROR";

const order: Record<Level, number> = { DEBUG: 0, INFO: 1, WARN: 2, ERROR: 3 };
const threshold = order[LOG_LEVEL.toUpperCase() as Level];

export const log = {
  debug: (msg: string, attrs?: Attrs) => write("DEBUG", msg, attrs),
  info: (msg: string, attrs?: Attrs) => write("INFO", msg, attrs),
  warn: (msg: string, attrs?: Attrs) => write("WARN", msg, attrs),
  error: (msg: string, attrs?: Attrs) => write("ERROR", msg, attrs),
  enabled: (level: Level) => order[level] >= threshold,
};

function write(level: Level, msg: string, attrs: Attrs = {}) {
  if (order[level] < threshold) return;
  const time = new Date().toISOString();
  const clean = normalize(attrs) as Attrs;
  const line =
    LOG_FORMAT === "json"
      ? JSON.stringify({ time, level, msg, ...clean })
      : [`time=${time}`, `level=${level}`, `msg=${quote(msg)}`, ...flatten(clean)].join(" ");
  (order[level] >= order.WARN ? process.stderr : process.stdout).write(line + "\n");
}

function normalize(value: unknown): unknown {
  if (value instanceof Error) return value.message;
  if (Array.isArray(value)) return value.map(normalize);
  if (value !== null && typeof value === "object") {
    return Object.fromEntries(
      Object.entries(value)
        .filter(([, v]) => v !== undefined)
        .map(([k, v]) => [k, normalize(v)]),
    );
  }
  return value;
}

function flatten(value: unknown, prefix = ""): string[] {
  if (value !== null && typeof value === "object" && !Array.isArray(value)) {
    return Object.entries(value).flatMap(([k, v]) => flatten(v, prefix ? `${prefix}.${k}` : k));
  }
  return [`${prefix}=${quote(value)}`];
}

function quote(value: unknown): string {
  if (value === null || value === undefined) return "<nil>";
  const text =
    typeof value === "string"
      ? value
      : typeof value === "object"
        ? JSON.stringify(value)
        : String(value);
  return text === "" || /[\s="\u0000-\u001f]/.test(text) ? JSON.stringify(text) : text;
}
