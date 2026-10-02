// Port of ../../main.go: looks up Minecraft player IDs (Bedrock XUID, Java UUID, Floodgate UUID).

export type Line = [label: string, value: string];
type Fetch = (url: string) => Promise<Response>;

// Mojang paths are proxied same-origin (vercel.json / astro.config.mjs); GeyserMC allows CORS.
const MOJANG_API = "/mojang/api";
const MOJANG_SESSION = "/mojang/session";
const GEYSER = "https://api.geysermc.org/v2/xbox";
const UINT64_MAX = 2n ** 64n - 1n;
const ZERO16 = "0".repeat(16);

export async function run(cmd: string, arg: string, get: Fetch = fetch): Promise<Line[]> {
  switch (cmd) {
    case "bedrock":
      return bedrockInfo(arg, await bedrockXUID(arg, get));
    case "java":
      return javaProfile(`${MOJANG_API}/users/profiles/minecraft/${encodeURIComponent(arg)}`, `java player ${q(arg)}`, get);
    case "reverse": {
      const h = javaUUID(arg);
      if (h) return javaProfile(`${MOJANG_SESSION}/session/minecraft/profile/${h}`, `java uuid ${arg}`, get);
      const xuid = parseXUID(arg);
      return bedrockInfo(await bedrockGamertag(xuid, get), xuid);
    }
    default:
      throw new Error(`unknown command ${q(cmd)}`);
  }
}

// javaUUID returns the undashed hex of s if s is a UUID that is not a Floodgate UUID.
export function javaUUID(s: string): string | undefined {
  const h = s.replaceAll("-", "").toLowerCase();
  return /^[0-9a-f]{32}$/.test(h) && !h.startsWith(ZERO16) ? h : undefined;
}

function bedrockInfo(gamertag: string, xuid: bigint): Line[] {
  return [
    ["Gamertag", gamertag],
    ["XUID(DEC)", xuid.toString()],
    ["XUID(HEX)", xuid.toString(16)],
    ["Floodgate UUID", floodgateUUID(xuid)],
  ];
}

// parseXUID accepts a decimal XUID, a hex XUID, or a Floodgate UUID.
// ponytail: digits-only input is read as decimal; a digits-only hex XUID needs the 0x prefix.
export function parseXUID(s: string): bigint {
  if (s.includes("-") || s.length === 32) return parseFloodgateUUID(s);
  const lower = s.toLowerCase();
  const prefixed = lower.startsWith("0x");
  if (!prefixed && /^\d+$/.test(s) && BigInt(s) <= UINT64_MAX) return BigInt(s);
  const h = prefixed ? lower.slice(2) : lower;
  if (!/^[0-9a-f]{1,16}$/.test(h)) throw new Error(`invalid xuid ${q(s)}`);
  return BigInt("0x" + h);
}

// ponytail: GeyserMC API only knows gamertags that have joined a Geyser server; full coverage needs Xbox Live auth.
async function bedrockXUID(gamertag: string, get: Fetch): Promise<bigint> {
  const what = `bedrock player ${q(gamertag)}`;
  const body = await getText(`${GEYSER}/xuid/${encodeURIComponent(gamertag)}`, what, get);
  // Read the digits instead of JSON.parse: a JS number loses XUID precision above 2^53.
  const m = body.match(/"xuid"\s*:\s*(\d+)/);
  if (!m || BigInt(m[1]) === 0n) throw new Error(`${what} not found`);
  return BigInt(m[1]);
}

async function bedrockGamertag(xuid: bigint, get: Fetch): Promise<string> {
  const what = `bedrock xuid ${xuid}`;
  const body = await getText(`${GEYSER}/gamertag/${xuid}`, what, get);
  const gamertag = body && JSON.parse(body).gamertag;
  if (!gamertag) throw new Error(`${what} not found`);
  return gamertag;
}

// javaProfile fetches a Mojang profile; both endpoints return {"id", "name"}.
async function javaProfile(url: string, what: string, get: Fetch): Promise<Line[]> {
  const body = await getText(url, what, get);
  const res = body ? JSON.parse(body) : {};
  if (res.id?.length !== 32 || !res.name) throw new Error(`${what} not found`);
  return [
    ["Name", res.name],
    ["UUID", dashed(res.id)],
  ];
}

// floodgateUUID matches Floodgate's `new UUID(0, xuid)`.
export function floodgateUUID(xuid: bigint): string {
  return dashed(ZERO16 + xuid.toString(16).padStart(16, "0"));
}

// parseFloodgateUUID is the inverse of floodgateUUID.
export function parseFloodgateUUID(uuid: string): bigint {
  const h = uuid.replaceAll("-", "");
  if (!/^0{16}[0-9a-fA-F]{16}$/.test(h)) throw new Error(`invalid floodgate uuid ${q(uuid)}`);
  return BigInt("0x" + h.slice(16));
}

function dashed(h: string): string {
  return `${h.slice(0, 8)}-${h.slice(8, 12)}-${h.slice(12, 16)}-${h.slice(16, 20)}-${h.slice(20, 32)}`;
}

const q = (s: string) => JSON.stringify(s);

// getText returns the body of a 200 response, or "" for Mojang's not-found 204/404; callers detect the empty value.
async function getText(url: string, what: string, get: Fetch): Promise<string> {
  let resp: Response;
  try {
    resp = await get(url);
  } catch (e) {
    throw new Error(`${what}: ${(e as Error).message}`);
  }
  if (resp.status === 204 || resp.status === 404) return "";
  const body = await resp.text();
  if (resp.status !== 200) {
    // GeyserMC uses "message" (503 when the gamertag is not cached), Mojang uses "errorMessage".
    let e: { message?: string; errorMessage?: string } = {};
    try {
      e = JSON.parse(body);
    } catch {} // best effort: the status code alone is still reported
    throw new Error(`${what}: HTTP ${resp.status} ${e.message ?? ""}${e.errorMessage ?? ""}`.trimEnd());
  }
  return body;
}
