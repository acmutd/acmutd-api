import {
  APIKey,
  APIKeyStatus,
  AppConfig,
  CronLog,
  DailyStat,
  InstanceState,
  PromotionLog,
  RequestEvent,
  ScraperLog,
  ServerStateLog,
  User
} from "../types/models";

import {
  mockApiKeys,
  mockAppConfig,
  mockCronLogs,
  mockDailyStats,
  mockInstanceState,
  mockPromotionLogs,
  mockRecentRequests,
  mockScraperLogs,
  mockServerStateLogs,
  mockUsers
} from "./mockData";

let users = [...mockUsers];
let apiKeys = [...mockApiKeys];
let dailyStats = [...mockDailyStats];
let recentRequests = [...mockRecentRequests];
let appConfig = { ...mockAppConfig };
let scraperLogs = [...mockScraperLogs];
let cronLogs = [...mockCronLogs];
let serverStateLogs = [...mockServerStateLogs];
let promotionLogs = [...mockPromotionLogs];
let instanceState = { ...mockInstanceState };
const currentUserNum = 4;
let sessionUserUid = users[currentUserNum]?.uid ?? "";

function nowIso(): string {
  return new Date().toISOString();
}

function generateId(prefix: string): string {
  return `${prefix}_${Math.random().toString(36).slice(2, 10)}`;
}

function buildTokenValue(): string {
  return `acm_live_${Math.random().toString(36).slice(2, 18)}`;
}

function getCurrentUserLocal(): User {
  const user = users.find((u) => u.uid === sessionUserUid);
  return user ?? users[0];
}

export async function getCurrentUser(): Promise<User> {
   
  return getCurrentUserLocal();
}

export async function loginWithGoogle(): Promise<User> {
   
  const user = users[currentUserNum];
  user.lastLoginAt = nowIso();
  sessionUserUid = user.uid;
  return user;
}

export async function signOut(): Promise<void> {
}

export async function setSessionUser(uid: string): Promise<User> {
  sessionUserUid = uid;
  return getCurrentUserLocal();
}

export async function getAppConfig(): Promise<AppConfig> {
   
  return { ...appConfig };
}

export async function getMyApiKey(userId: string): Promise<APIKey | null> {
   
  return apiKeys.find((k) => k.userId === userId) ?? null;
}

export async function requestApiKey(label: string, description: string): Promise<void> {
   
  const user = getCurrentUserLocal();
  const pendingKey: APIKey = {
    keyId: generateId("key"),
    key: buildTokenValue(),
    userId: user.uid,
    ownerEmail: user.email,
    label,
    description,
    status: "pending",
    isAdmin: false,
    rateLimit: 60,
    windowSeconds: 60,
    createdAt: nowIso(),
    expiresAt: new Date(Date.now() + 1000 * 60 * 60 * 24 * 365).toISOString(),
    usageCount: 0,
    lastUsedAt: null
  };
  apiKeys = [...apiKeys.filter((k) => k.userId !== user.uid), pendingKey];
}

export async function regenerateApiKey(keyId: string): Promise<APIKey> {
   
  const key = apiKeys.find((k) => k.keyId === keyId);
  if (!key) {
    throw new Error("Key not found");
  }
  key.key = buildTokenValue();
  key.status = "active";
  key.lastUsedAt = null;
  return { ...key };
}

export async function revokeApiKey(keyId: string): Promise<void> {
   
  apiKeys = apiKeys.map((key) =>
    key.keyId === keyId ? { ...key, status: "inactive" as APIKeyStatus } : key
  );
}

export async function getMyDailyStats(userId: string): Promise<DailyStat[]> {
   
  return dailyStats.filter((s) => s.userId === userId);
}

export async function getMyRecentRequests(userId: string): Promise<RequestEvent[]> {
   
  const myKeyIds = new Set(apiKeys.filter((k) => k.userId === userId).map((k) => k.keyId));
  if (!myKeyIds.size) {
    return [];
  }
  return recentRequests.filter(
    (request) => request.userId === userId || myKeyIds.has(request.keyId)
  );
}

export async function listAllDailyStats(): Promise<DailyStat[]> {
  return dailyStats.map((stat) => ({ ...stat }));
}

export async function listAllRecentRequests(): Promise<RequestEvent[]> {
  return recentRequests.map((request) => ({ ...request }));
}

export async function listAllKeys(): Promise<APIKey[]> {
   
  return [...apiKeys];
}

export async function listPendingKeyRequests(): Promise<APIKey[]> {
   
  return apiKeys.filter((k) => k.status === "pending");
}

export async function approveKey(keyId: string): Promise<void> {
   
  apiKeys = apiKeys.map((key) =>
    key.keyId === keyId
      ? {
          ...key,
          status: "active",
          createdAt: key.createdAt || nowIso()
        }
      : key
  );
}

export async function rejectKey(keyId: string): Promise<void> {
   
  apiKeys = apiKeys.map((key) =>
    key.keyId === keyId
      ? {
          ...key,
          status: "rejected"
        }
      : key
  );
}

export async function adminRegenerateKey(keyId: string): Promise<APIKey> {
  return regenerateApiKey(keyId);
}

export async function adminRevokeKey(keyId: string): Promise<void> {
  return revokeApiKey(keyId);
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
   
  let owner = users.find((u) => u.email.toLowerCase() === input.ownerEmail.toLowerCase());
  if (!owner) {
    owner = {
      uid: generateId("u"),
      email: input.ownerEmail,
      displayName: input.ownerEmail.split("@")[0],
      isAdmin: input.isAdmin,
      approvalStatus: "approved",
      createdAt: nowIso(),
      lastLoginAt: nowIso()
    };
    users = [...users, owner];
  }

  const token: APIKey = {
    keyId: generateId("key"),
    key: buildTokenValue(),
    userId: owner.uid,
    ownerEmail: owner.email,
    label: input.label,
    description: input.description,
    status: "active",
    isAdmin: input.isAdmin,
    rateLimit: input.rateLimit,
    windowSeconds: input.windowSeconds,
    createdAt: nowIso(),
    expiresAt: input.expiresAt,
    usageCount: 0,
    lastUsedAt: null
  };

  apiKeys = [...apiKeys, token];
  return token;
}

export async function editToken(
  keyId: string,
  update: Partial<Pick<APIKey, "ownerEmail" | "label" | "description" | "isAdmin" | "rateLimit" | "windowSeconds" | "expiresAt">>
): Promise<APIKey> {
   
  const target = apiKeys.find((k) => k.keyId === keyId);
  if (!target) {
    throw new Error("Token not found");
  }

  if (update.ownerEmail) {
    target.ownerEmail = update.ownerEmail;
  }
  if (update.label) {
    target.label = update.label;
  }
  if (update.description) {
    target.description = update.description;
  }
  if (typeof update.isAdmin === "boolean") {
    target.isAdmin = update.isAdmin;
  }
  if (typeof update.rateLimit === "number") {
    target.rateLimit = update.rateLimit;
  }
  if (typeof update.windowSeconds === "number") {
    target.windowSeconds = update.windowSeconds;
  }
  if (update.expiresAt) {
    target.expiresAt = update.expiresAt;
  }

  return { ...target };
}

export async function deactivateToken(keyId: string): Promise<void> {
  await revokeApiKey(keyId);
}

export async function listUsers(): Promise<User[]> {
  return users.map((user) => ({ ...user }));
}

export async function banUser(uid: string): Promise<void> {
  users = users.map((user) =>
    user.uid === uid ? { ...user, approvalStatus: "banned" } : user
  );
  apiKeys = apiKeys.map((key) =>
    key.userId === uid ? { ...key, status: "inactive" as APIKeyStatus } : key
  );
}

export async function unbanUser(uid: string): Promise<void> {
  users = users.map((user) =>
    user.uid === uid ? { ...user, approvalStatus: "approved" } : user
  );
}

export async function listScraperLogs(): Promise<ScraperLog[]> {
   
  return [...scraperLogs];
}

export async function listCronLogs(): Promise<CronLog[]> {
   
  return [...cronLogs];
}

export async function listServerStateLogs(): Promise<ServerStateLog[]> {
  return [...serverStateLogs];
}

export async function listPromotionLogs(): Promise<PromotionLog[]> {
   
  return [...promotionLogs];
}

export async function getInstanceState(): Promise<InstanceState> {
   
  return { ...instanceState };
}

type ChangeServerStateInput = {
  action: ServerStateLog["action"];
  triggerSource: ServerStateLog["triggerSource"];
  actorType: ServerStateLog["actorType"];
  actorId: string;
  actorEmail?: string;
  reason: string;
  cronLogId?: string;
  nextState?: InstanceState["state"];
  nextType?: InstanceState["instanceType"];
  forceError?: boolean;
};

async function changeServerState(input: ChangeServerStateInput): Promise<ServerStateLog> {
  const startedAt = Date.now();
  const previousState = instanceState.state;
  const previousType = instanceState.instanceType;
  const requestId = generateId("req");

  const targetState = input.nextState ?? previousState;
  const targetType = input.nextType ?? previousType;

  if (!input.forceError) {
    instanceState = {
      ...instanceState,
      state: targetState,
      instanceType: targetType,
      uptimeSeconds: targetState === "running" ? 0 : instanceState.uptimeSeconds
    };
  }

  const log: ServerStateLog = {
    logId: generateId("srv"),
    timestamp: nowIso(),
    action: input.action,
    status: input.forceError ? "error" : "success",
    triggerSource: input.triggerSource,
    actorType: input.actorType,
    actorId: input.actorId,
    actorEmail: input.actorEmail ?? "",
    reason: input.reason,
    requestId,
    cronLogId: input.cronLogId ?? "",
    previousState,
    newState: input.forceError ? previousState : targetState,
    previousType,
    newType: input.forceError ? previousType : targetType,
    instanceId: instanceState.instanceId,
    publicIp: instanceState.publicIp,
    errorMessage: input.forceError ? "Simulated EC2 API failure" : "",
    durationMs: Date.now() - startedAt
  };

  serverStateLogs = [log, ...serverStateLogs];
  return log;
}

export async function startServer(): Promise<void> {
  const actor = getCurrentUserLocal();
  await changeServerState({
    action: "start",
    triggerSource: "admin_dashboard",
    actorType: "user",
    actorId: actor.uid,
    actorEmail: actor.email,
    reason: "Manual start requested from admin dashboard",
    nextState: "running"
  });
}

export async function stopServer(): Promise<void> {
  const actor = getCurrentUserLocal();

  if (appConfig.hackathonModeEnabled) {
    const fallbackType: InstanceState["instanceType"] = "t3.micro";
    await changeServerState({
      action: "hackathon_disable",
      triggerSource: "admin_dashboard",
      actorType: "user",
      actorId: actor.uid,
      actorEmail: actor.email,
      reason: "Hackathon mode automatically disabled during manual stop",
      nextType: fallbackType
    });
    await updateAppConfig({
      hackathonModeEnabled: false,
      instanceType: fallbackType
    });
  }

  await changeServerState({
    action: "stop",
    triggerSource: "admin_dashboard",
    actorType: "user",
    actorId: actor.uid,
    actorEmail: actor.email,
    reason: "Manual stop requested from admin dashboard",
    nextState: "stopped"
  });
}

export async function updateAppConfig(cfg: Partial<AppConfig>): Promise<AppConfig> {
   
  appConfig = {
    ...appConfig,
    ...cfg,
    updatedAt: nowIso(),
    updatedBy: getCurrentUserLocal().uid
  };

  instanceState = {
    ...instanceState,
    instanceType: appConfig.instanceType
  };
  return { ...appConfig };
}

export async function enableHackathonMode(): Promise<void> {
  const actor = getCurrentUserLocal();
  const nextType: InstanceState["instanceType"] = "t3.large";
  await changeServerState({
    action: "hackathon_enable",
    triggerSource: "admin_dashboard",
    actorType: "user",
    actorId: actor.uid,
    actorEmail: actor.email,
    reason: "Hackathon mode enabled from admin dashboard",
    nextType
  });
  await updateAppConfig({ hackathonModeEnabled: true, instanceType: nextType });
}

export async function disableHackathonMode(): Promise<void> {
  const actor = getCurrentUserLocal();
  const nextType: InstanceState["instanceType"] = "t3.micro";
  await changeServerState({
    action: "hackathon_disable",
    triggerSource: "admin_dashboard",
    actorType: "user",
    actorId: actor.uid,
    actorEmail: actor.email,
    reason: "Hackathon mode disabled from admin dashboard",
    nextType
  });
  await updateAppConfig({ hackathonModeEnabled: false, instanceType: nextType });
}
