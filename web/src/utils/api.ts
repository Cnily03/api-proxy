import ky, { type KyInstance } from "ky";

function resolveBase(): string {
  const raw = (import.meta.env.VITE_API_BASE as string | undefined) ?? "/api";
  if (/^https?:\/\//.test(raw)) return raw.replace(/\/$/, "");
  return `${window.location.origin}${raw.startsWith("/") ? "" : "/"}${raw}`.replace(/\/$/, "");
}

export class APIError extends Error {
  response: Response | null;
  request: Request | null;
  constructor(message: string, { response, request }: { response?: Response; request?: Request }) {
    super(message);
    this.response = response ?? null;
    this.request = request ?? null;
    this.name = "APIError";
  }
}

const api: KyInstance = ky.create({
  prefix: resolveBase(),
  hooks: {
    beforeRequest: [
      ({ request }) => {
        const token = localStorage.getItem("token");
        if (token) {
          request.headers.set("Authorization", `Bearer ${token}`);
        }
      },
    ],
    afterResponse: [
      async ({ response, request }) => {
        if (!response.ok) {
          const { status, statusText } = response;
          let body = "";
          try {
            body = (await response.clone().text()).trim();
          } catch {}
          throw new APIError(
            body ? `请求失败：${status} ${statusText}: ${body}` : `请求失败：${status} ${statusText}`,
            { response, request }
          );
        }
        // Auto-save token from X-Set-Token header (login & renewal)
        const newToken = response.headers.get("X-Set-Token");
        if (newToken) {
          localStorage.setItem("token", newToken);
        }
        return response;
      },
    ],
  },
});

// ── Crypto helper ──
async function sha256Hex(message: string): Promise<string> {
  const data = new TextEncoder().encode(message);
  const hashBuffer = await crypto.subtle.digest("SHA-256", data);
  return Array.from(new Uint8Array(hashBuffer))
    .map((b) => b.toString(16).padStart(2, "0"))
    .join("");
}

const PASSWORD_SALT = (import.meta.env.VITE_PASSWORD_SALT as string | undefined) ?? ".openicu";

export function hashPassword(plain: string): Promise<string> {
  return sha256Hex(plain + PASSWORD_SALT);
}

// ── Auth ──
export async function login(username: string, password: string): Promise<void> {
  const hashed = await hashPassword(password);
  await api.post("auth/login", { json: { username, password: hashed } });
}

export function getMe() {
  return api.get("auth/me").json<{ id: number; username: string; role: string }>();
}

export async function changePassword(oldPassword: string, newPassword: string) {
  const [oldHash, newHash] = await Promise.all([hashPassword(oldPassword), hashPassword(newPassword)]);
  return api
    .post("auth/password", { json: { old_password: oldHash, new_password: newHash } })
    .json<{ status: string }>();
}

// ── Proxy ──
export function getProxyEndpoint() {
  return api.get("proxy/endpoint").json<{ endpoint: string }>();
}

// ── Rules ──
export type Rule = {
  id: number;
  name: string;
  src: string;
  dest: string;
  dest_api_key: string;
  skip_cert_verify: boolean;
  api_key: string;
  force_api_key: boolean;
  comment: string;
  tags: string[];
};

export function listRules() {
  return api.get("rules").json<Rule[]>();
}

export function createRule(rule: Omit<Rule, "id">) {
  return api.post("rules", { json: rule }).json<{ id: number }>();
}

export function updateRule(rule: Rule) {
  return api.put(`rules/${rule.id}`, { json: rule }).json<{ status: string }>();
}

export function deleteRule(id: number) {
  return api.delete(`rules/${id}`).json<{ status: string }>();
}

// ── Users ──
export type User = {
  id: number;
  username: string;
  role: "admin" | "user";
};

export function listUsers() {
  return api.get("users").json<User[]>();
}

export async function createUser(username: string, password: string, role: string) {
  const hashed = await hashPassword(password);
  return api.post("users", { json: { username, password: hashed, role } }).json<{ id: number }>();
}

export async function updateUser(id: number, data: { username?: string; password?: string; role?: string }) {
  const payload = { ...data };
  if (payload.password) {
    payload.password = await hashPassword(payload.password);
  }
  return api.put(`users/${id}`, { json: payload }).json<{ status: string }>();
}

export function deleteUser(id: number) {
  return api.delete(`users/${id}`).json<{ status: string }>();
}
