import {
  APIKey,
  AppConfig,
  CronLog,
  DailyStat,
  InstanceState,
  PromotionLog,
  RequestEvent,
  ScraperLog,
  ServerStateLog,
  User,
} from "../types/models";
import { firebaseAuth } from "./firebase";

const BASE = "/dashboard/api/v1";

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = await firebaseAuth.currentUser?.getIdToken();
  if (!token) throw new Error("not authenticated");

  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      "Content-Type": "application/json",
      Authorization: `Bearer ${token}`,
      ...(init?.headers ?? {}),
    },
  });

  if (!res.ok) {
    const body = await res.json().catch(() => ({})) as Record<string, string>;
    throw new Error(body.error ?? `HTTP ${res.status}`);
  }

  return res.json() as Promise<T>;
}

// ---------------------------------------------------------------------------
// Auth helpers — actual sign-in/sign-out is handled by AuthContext via the
// Firebase SDK directly. These stubs satisfy the shared function-signature
// contract with mockApi.ts so apiClient.ts can re-export either file.
// ---------------------------------------------------------------------------

export async function getCurrentUser(): Promise<User | null> {
  return null;
}

export async function loginWithGoogle(): Promise<User> {
  throw new Error("loginWithGoogle must be called through AuthContext");
}

export async function signOut(): Promise<void> {
  return firebaseAuth.signOut();
}

// No-op in production — only meaningful when apiClient.ts points to mockApi.ts
export async function setSessionUser(_uid: string): Promise<User> {
  throw new Error("setSessionUser is only available in mock mode");
}

// ---------------------------------------------------------------------------
// App config
// ---------------------------------------------------------------------------

export async function getAppConfig(): Promise<AppConfig> {
  return apiFetch<AppConfig>("/config");
}

export async function updateAppConfig(cfg: Partial<AppConfig>): Promise<AppConfig> {
  return apiFetch<AppConfig>("/admin/config", {
    method: "PUT",
    body: JSON.stringify(cfg),
  });
}

// ---------------------------------------------------------------------------
// User-facing: API key management
// ---------------------------------------------------------------------------

export async function getMyApiKey(_userId: string): Promise<APIKey | null> {
  return apiFetch<APIKey | null>("/keys/me");
}

export async function requestApiKey(label: string, description: string): Promise<void> {
  await apiFetch("/keys/request", {
    method: "POST",
    body: JSON.stringify({ label, description }),
  });
}

export async function regenerateApiKey(keyId: string): Promise<APIKey> {
  return apiFetch<APIKey>(`/keys/${keyId}/regenerate`, { method: "POST" });
}

export async function revokeApiKey(keyId: string): Promise<void> {
  await apiFetch(`/keys/${keyId}`, { method: "DELETE" });
}

// ---------------------------------------------------------------------------
// User-facing: usage stats
// ---------------------------------------------------------------------------

export async function getMyDailyStats(_userId: string): Promise<DailyStat[]> {
  const res = await apiFetch<{ stats: DailyStat[] }>("/stats/daily");
  return res.stats ?? [];
}

export async function getMyRecentRequests(_userId: string): Promise<RequestEvent[]> {
  const res = await apiFetch<{ requests: RequestEvent[] }>("/stats/requests");
  return res.requests ?? [];
}

// ---------------------------------------------------------------------------
// Admin: key management
// ---------------------------------------------------------------------------

export async function listAllKeys(): Promise<APIKey[]> {
  const res = await apiFetch<{ keys: APIKey[] }>("/admin/keys");
  return res.keys ?? [];
}

export async function listPendingKeyRequests(): Promise<APIKey[]> {
  const res = await apiFetch<{ keys: APIKey[] }>("/admin/keys/pending");
  return res.keys ?? [];
}

export async function approveKey(keyId: string): Promise<void> {
  await apiFetch(`/admin/keys/${keyId}/approve`, { method: "POST" });
}

export async function rejectKey(keyId: string): Promise<void> {
  await apiFetch(`/admin/keys/${keyId}/reject`, { method: "POST" });
}

export async function adminRegenerateKey(keyId: string): Promise<APIKey> {
  return apiFetch<APIKey>(`/admin/keys/${keyId}/regenerate`, { method: "POST" });
}

export async function adminRevokeKey(keyId: string): Promise<void> {
  await apiFetch(`/admin/keys/${keyId}`, { method: "DELETE" });
}

export async function addToken(input: {
  ownerEmail: string;
  label: string;
  description: string;
  isAdmin: boolean;
  rateLimit: number;
  windowSeconds: number;
  expiresAt: string;
}): Promise<APIKey> {
  return apiFetch<APIKey>("/admin/keys", {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export async function editToken(
  keyId: string,
  update: Partial<
    Pick<APIKey, "ownerEmail" | "label" | "description" | "isAdmin" | "rateLimit" | "windowSeconds" | "expiresAt">
  >
): Promise<APIKey> {
  return apiFetch<APIKey>(`/admin/keys/${keyId}`, {
    method: "PUT",
    body: JSON.stringify(update),
  });
}

export async function deactivateToken(keyId: string): Promise<void> {
  await apiFetch(`/admin/keys/${keyId}`, { method: "DELETE" });
}

// ---------------------------------------------------------------------------
// Admin: usage stats
// ---------------------------------------------------------------------------

export async function listAllDailyStats(): Promise<DailyStat[]> {
  const res = await apiFetch<{ stats: DailyStat[] }>("/admin/stats/daily");
  return res.stats ?? [];
}

export async function listAllRecentRequests(): Promise<RequestEvent[]> {
  const res = await apiFetch<{ requests: RequestEvent[] }>("/admin/stats/requests");
  return res.requests ?? [];
}

// ---------------------------------------------------------------------------
// Admin: user management
// ---------------------------------------------------------------------------

export async function listUsers(): Promise<User[]> {
  const res = await apiFetch<{ users: User[] }>("/users");
  return res.users ?? [];
}

export async function banUser(uid: string): Promise<void> {
  await apiFetch(`/users/${uid}/ban`, { method: "PUT" });
}

export async function unbanUser(uid: string): Promise<void> {
  await apiFetch(`/users/${uid}/unban`, { method: "PUT" });
}

// ---------------------------------------------------------------------------
// Admin: server controls
// ---------------------------------------------------------------------------

export async function getInstanceState(): Promise<InstanceState> {
  return apiFetch<InstanceState>("/admin/server/state");
}

export async function startServer(): Promise<void> {
  await apiFetch("/admin/server/start", { method: "POST" });
}

export async function stopServer(): Promise<void> {
  await apiFetch("/admin/server/stop", { method: "POST" });
}

export async function enableHackathonMode(): Promise<void> {
  await apiFetch("/admin/server/hackathon/enable", { method: "POST" });
}

export async function disableHackathonMode(): Promise<void> {
  await apiFetch("/admin/server/hackathon/disable", { method: "POST" });
}

// ---------------------------------------------------------------------------
// Admin: logs
// ---------------------------------------------------------------------------

export async function listScraperLogs(): Promise<ScraperLog[]> {
  const res = await apiFetch<{ logs: ScraperLog[] }>("/admin/logs/scraper");
  return res.logs ?? [];
}

export async function listCronLogs(): Promise<CronLog[]> {
  const res = await apiFetch<{ logs: CronLog[] }>("/admin/logs/cron");
  return res.logs ?? [];
}

export async function listServerStateLogs(): Promise<ServerStateLog[]> {
  const res = await apiFetch<{ logs: ServerStateLog[] }>("/admin/logs/server");
  return res.logs ?? [];
}

export async function listPromotionLogs(): Promise<PromotionLog[]> {
  const res = await apiFetch<{ logs: PromotionLog[] }>("/admin/logs/promotion");
  return res.logs ?? [];
}
