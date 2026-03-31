import { useEffect, useMemo, useState } from "react";
import { PortalLayout } from "../components/layout/PortalLayout";
import { StatusBadge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { Card, CardSubtitle, CardTitle } from "../components/ui/Card";
import { Input, Label, Select, TextArea } from "../components/ui/Field";
import { Modal } from "../components/ui/Modal";
import {
  addToken,
  adminRegenerateKey,
  adminRevokeKey,
  approveKey,
  banUser,
  deactivateToken,
  disableHackutdMode,
  editToken,
  enableHackutdMode,
  getAppConfig,
  getInstanceState,
  listAllKeys,
  listAllDailyStats,
  listAllRecentRequests,
  listCronLogs,
  listPendingKeyRequests,
  listPromotionLogs,
  listScraperLogs,
  listUsers,
  rejectKey,
  startServer,
  stopServer,
  unbanUser,
  updateAppConfig,
} from "../lib/api";
import {
  formatDate,
  formatDateTime,
  formatPercent,
  formatUptime,
  maskKey,
} from "../lib/format";
import { useAuth } from "../state/AuthContext";
import { useToast } from "../state/ToastContext";
import {
  APIKey,
  AppConfig,
  CronLog,
  DailyStat,
  InstanceState,
  PromotionLog,
  RequestEvent,
  ScraperLog,
  User,
} from "../types/models";

const adminNavItems = [
  { key: "overview", label: "Overview" },
  { key: "user-keys", label: "User Keys" },
  { key: "token-management", label: "Usage" },
  { key: "server-controls", label: "Server Controls" },
  { key: "logs-metrics", label: "Logs & Metrics" },
  { key: "config", label: "Config" },
];

type AdminTab = (typeof adminNavItems)[number]["key"];

export function AdminDashboardPage() {
  const { currentUser, logout } = useAuth();
  const { showToast } = useToast();

  const [activeTab, setActiveTab] = useState<AdminTab>("overview");
  const [appConfig, setAppConfig] = useState<AppConfig | null>(null);
  const [allKeys, setAllKeys] = useState<APIKey[]>([]);
  const [pendingKeys, setPendingKeys] = useState<APIKey[]>([]);
  const [scraperLogs, setScraperLogs] = useState<ScraperLog[]>([]);
  const [cronLogs, setCronLogs] = useState<CronLog[]>([]);
  const [promotionLogs, setPromotionLogs] = useState<PromotionLog[]>([]);
  const [allDailyStats, setAllDailyStats] = useState<DailyStat[]>([]);
  const [allRecentRequests, setAllRecentRequests] = useState<RequestEvent[]>(
    [],
  );
  const [instanceState, setInstanceState] = useState<InstanceState | null>(
    null,
  );
  const [loading, setLoading] = useState(true);

  const [statusFilter, setStatusFilter] = useState("all");
  const [emailSearch, setEmailSearch] = useState("");
  const [usageUserFilter, setUsageUserFilter] = useState("all");

  const [tokenModalOpen, setTokenModalOpen] = useState(false);
  const [tokenModalMode, setTokenModalMode] = useState<"add" | "edit">("add");
  const [editingTokenId, setEditingTokenId] = useState<string | null>(null);
  const [tokenForm, setTokenForm] = useState({
    ownerEmail: "",
    label: "",
    description: "",
    isAdmin: false,
    rateLimit: 120,
    windowSeconds: 60,
    expiresAt: "2026-12-31",
  });
  const [usersById, setUsersById] = useState<Record<string, User>>({});
  const [userModerationModalOpen, setUserModerationModalOpen] = useState(false);
  const [moderationTarget, setModerationTarget] = useState<{
    uid: string;
    email: string;
    action: "ban" | "unban";
  } | null>(null);

  const [configDraft, setConfigDraft] = useState<AppConfig | null>(null);

  const filteredKeys = useMemo(() => {
    return allKeys.filter((key) => {
      const matchesStatus =
        statusFilter === "all" ? true : key.status === statusFilter;
      const matchesEmail = key.ownerEmail
        .toLowerCase()
        .includes(emailSearch.toLowerCase());
      return matchesStatus && matchesEmail;
    });
  }, [allKeys, statusFilter, emailSearch]);

  const activeKeyCount = allKeys.filter((k) => k.status === "active").length;
  const todayRequests = allKeys.reduce(
    (sum, key) => sum + Math.min(key.usageCount, 500),
    0,
  );
  const todayErrors = Math.max(Math.round(todayRequests * 0.02), 1);

  const filteredUsageStats = useMemo(() => {
    return allDailyStats.filter((stat) =>
      usageUserFilter === "all" ? true : stat.userId === usageUserFilter,
    );
  }, [allDailyStats, usageUserFilter]);

  const filteredUsageRequests = useMemo(() => {
    return allRecentRequests.filter((request) =>
      usageUserFilter === "all" ? true : request.userId === usageUserFilter,
    );
  }, [allRecentRequests, usageUserFilter]);

  async function loadData() {
    setLoading(true);
    const [
      cfg,
      keys,
      pending,
      scraper,
      cron,
      promotions,
      dailyStats,
      recentRequests,
      instance,
      users,
    ] = await Promise.all([
      getAppConfig(),
      listAllKeys(),
      listPendingKeyRequests(),
      listScraperLogs(),
      listCronLogs(),
      listPromotionLogs(),
      listAllDailyStats(),
      listAllRecentRequests(),
      getInstanceState(),
      listUsers(),
    ]);

    setAppConfig(cfg);
    setConfigDraft(cfg);
    setAllKeys(keys);
    setPendingKeys(pending);
    setScraperLogs(scraper);
    setCronLogs(cron);
    setPromotionLogs(promotions);
    setAllDailyStats(dailyStats);
    setAllRecentRequests(recentRequests);
    setInstanceState(instance);
    setUsersById(
      users.reduce<Record<string, User>>((acc, user) => {
        acc[user.uid] = user;
        return acc;
      }, {}),
    );
    setLoading(false);
  }

  useEffect(() => {
    void loadData();
  }, []);

  if (!currentUser || !currentUser.isAdmin) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50 p-6">
        <div className="rounded-xl border border-slate-200 bg-white p-8 text-center">
          <h1 className="text-xl font-bold">Admin access required</h1>
          <p className="mt-2 text-sm text-slate-600">
            Use a mock admin account to view this page.
          </p>
        </div>
      </div>
    );
  }

  if (loading || !appConfig || !instanceState || !configDraft) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <p className="text-sm text-slate-600">Loading admin dashboard...</p>
      </div>
    );
  }

  return (
    <PortalLayout
      user={currentUser}
      appConfig={appConfig}
      title="Admin Dashboard"
      navItems={adminNavItems}
      activeTab={activeTab}
      onTabChange={(tab) => setActiveTab(tab as AdminTab)}
      onSignOut={logout}
    >
      {activeTab === "overview" && (
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-4">
            <Card>
              <CardSubtitle>Total Active Keys</CardSubtitle>
              <CardTitle>{activeKeyCount}</CardTitle>
            </Card>
            <Card>
              <CardSubtitle>Requests Today</CardSubtitle>
              <CardTitle>{todayRequests.toLocaleString()}</CardTitle>
            </Card>
            <Card>
              <CardSubtitle>Error Rate (today)</CardSubtitle>
              <CardTitle>
                {((todayErrors / todayRequests) * 100).toFixed(2)}%
              </CardTitle>
            </Card>
            <Card>
              <CardSubtitle>Instance Type</CardSubtitle>
              <CardTitle>{appConfig.instanceType}</CardTitle>
            </Card>
          </div>

          <Card>
            <CardTitle>HackUTD Mode</CardTitle>
            <CardSubtitle>
              Operational switch tied to infrastructure profile.
            </CardSubtitle>
            <div className="mt-4 flex flex-wrap items-center gap-3">
              <span
                className={`rounded-full px-3 py-1 text-xs font-semibold ${appConfig.hackutdModeEnabled ? "bg-amber-100 text-amber-700" : "bg-slate-100 text-slate-700"}`}
              >
                {appConfig.hackutdModeEnabled ? "Enabled" : "Disabled"}
              </span>
              <Button
                variant="secondary"
                disabled
                title="Disabled in skeleton; wired to mock updateAppConfig."
              >
                Toggle (disabled)
              </Button>
              <p className="text-sm text-slate-600">
                End date: {formatDateTime(appConfig.hackutdEndDate)}
              </p>
            </div>
          </Card>
        </div>
      )}

      {activeTab === "user-keys" && (
        <div className="flex h-full min-h-0 flex-col gap-4">
          <Card>
            <CardTitle>Approval Queue</CardTitle>
            <CardSubtitle>
              Pending API key requests awaiting moderation.
            </CardSubtitle>
            <div className="mt-3 space-y-2">
              {pendingKeys.length === 0 && (
                <p className="text-sm text-slate-500">No pending requests.</p>
              )}
              {pendingKeys.map((key) => (
                <div
                  key={key.keyId}
                  className="flex flex-col gap-2 rounded-lg border border-slate-200 p-3 md:flex-row md:items-center md:justify-between"
                >
                  <div>
                    <p className="text-sm font-semibold">{key.ownerEmail}</p>
                    <p className="text-sm text-slate-600">
                      {key.label} - {key.description}
                    </p>
                    <p className="text-xs text-slate-500">
                      Requested: {formatDateTime(key.createdAt)}
                    </p>
                  </div>
                  <div className="flex gap-2">
                    <Button
                      onClick={async () => {
                        await approveKey(key.keyId);
                        showToast("Key approved");
                        await loadData();
                      }}
                    >
                      Approve
                    </Button>
                    <Button
                      variant="danger"
                      onClick={async () => {
                        await rejectKey(key.keyId);
                        showToast("Key rejected");
                        await loadData();
                      }}
                    >
                      Reject
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          </Card>

          <Card className="flex min-h-0 flex-1 flex-col">
            <div className="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <CardTitle>User Keys</CardTitle>
              <Button
                onClick={() => {
                  setTokenModalMode("add");
                  setEditingTokenId(null);
                  setTokenForm({
                    ownerEmail: "",
                    label: "",
                    description: "",
                    isAdmin: false,
                    rateLimit: 120,
                    windowSeconds: 60,
                    expiresAt: "2026-12-31",
                  });
                  setTokenModalOpen(true);
                }}
                className="grid gap-2 md:grid-cols-2"
              >
                Add key
              </Button>
              <div className="grid gap-2 md:grid-cols-2">
                <Select
                  value={statusFilter}
                  onChange={(e) => setStatusFilter(e.target.value)}
                >
                  <option value="all">All statuses</option>
                  <option value="active">Active</option>
                  <option value="pending">Pending</option>
                  <option value="inactive">Inactive</option>
                  <option value="rejected">Rejected</option>
                </Select>
                <Input
                  value={emailSearch}
                  onChange={(e) => setEmailSearch(e.target.value)}
                  placeholder="Search owner email"
                />
              </div>
            </div>

            {/*
              Individual key info
            */}

            <div className="min-h-0 flex-1 overflow-auto">
              <table className="min-w-full text-left text-sm">
                <thead className="border-b border-slate-200 text-slate-500">
                  <tr>
                    <th className="py-2 pr-4">Label</th>
                    <th className="py-2 pr-4">Owner email</th>
                    <th className="py-2 pr-4">Status</th>
                    <th className="py-2 pr-4">Type</th>
                    <th className="py-2 pr-4">Created</th>
                    <th className="py-2 pr-4">Expires</th>
                    <th className="py-2 pr-4">Usage</th>
                    <th className="py-2 pr-4">Last used</th>
                    <th className="py-2 pr-4">Actions</th>
                  </tr>
                </thead>

                <tbody>
                  {filteredKeys.map((key) => (
                    <tr key={key.keyId} className="border-b border-slate-100">
                      <td className="py-2 pr-4">
                        <div>
                          <p>{key.label}</p>
                          <div className="group relative mt-1 inline-block">
                            <button
                              type="button"
                              className="font-mono text-xs text-slate-500 underline decoration-dotted underline-offset-2"
                              onClick={async () => {
                                try {
                                  await navigator.clipboard.writeText(key.key);
                                  showToast("API key copied to clipboard");
                                } catch {
                                  showToast("Unable to copy key");
                                }
                              }}
                            >
                              {maskKey(key.key)}
                            </button>
                            <div className="pointer-events-none invisible absolute left-0 top-full z-20 mt-1 w-80 rounded-md border border-slate-200 bg-white p-2 font-mono text-xs text-slate-700 opacity-0 shadow-md transition group-hover:visible group-hover:opacity-100">
                              <p>{key.key}</p>
                              <p className="mt-1 font-sans text-[11px] text-slate-500">
                                Click masked key to copy
                              </p>
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="py-2 pr-4">
                        <div className="group relative inline-flex">
                          <span className="cursor-help underline decoration-dotted underline-offset-2">
                            {key.ownerEmail}
                          </span>
                          <div className="invisible absolute left-0 top-full z-20 mt-2 w-72 rounded-md border border-slate-200 bg-white p-3 text-xs text-slate-700 opacity-0 shadow-md transition group-hover:visible group-hover:opacity-100 group-focus-within:visible group-focus-within:opacity-100">
                            <p className="text-sm font-semibold text-slate-900">
                              {usersById[key.userId]?.displayName ||
                                "Unknown user"}
                            </p>
                            <p className="mt-1">Email: {key.ownerEmail}</p>
                            <p className="mt-1">UID: {key.userId}</p>
                            <p className="mt-1">
                              Role:{" "}
                              {usersById[key.userId]?.isAdmin
                                ? "Admin"
                                : "User"}
                            </p>
                            <p className="mt-1">
                              Status:{" "}
                              {usersById[key.userId]?.approvalStatus ||
                                "unknown"}
                            </p>
                            <p className="mt-1">
                              Last login:{" "}
                              {formatDateTime(
                                usersById[key.userId]?.lastLoginAt || null,
                              )}
                            </p>
                            <div className="mt-3">
                              {usersById[key.userId]?.isAdmin ? (
                                <Button
                                  variant="danger"
                                  className="w-full"
                                  disabled
                                >
                                  Cannot ban admin
                                </Button>
                              ) : usersById[key.userId]?.approvalStatus ===
                                "banned" ? (
                                <Button
                                  variant="secondary"
                                  className="w-full"
                                  onClick={() => {
                                    setModerationTarget({
                                      uid: key.userId,
                                      email: key.ownerEmail,
                                      action: "unban",
                                    });
                                    setUserModerationModalOpen(true);
                                  }}
                                >
                                  Unban user
                                </Button>
                              ) : (
                                <Button
                                  variant="danger"
                                  className="w-full"
                                  onClick={() => {
                                    setModerationTarget({
                                      uid: key.userId,
                                      email: key.ownerEmail,
                                      action: "ban",
                                    });
                                    setUserModerationModalOpen(true);
                                  }}
                                >
                                  Ban user
                                </Button>
                              )}
                            </div>
                          </div>
                        </div>
                      </td>
                      <td className="py-2 pr-4">
                        <StatusBadge status={key.status} />
                      </td>
                      <td className="py-2 pr-4">
                        {key.isAdmin ? "Admin" : "Normal"}
                      </td>
                      <td className="py-2 pr-4">{formatDate(key.createdAt)}</td>
                      <td className="py-2 pr-4">{formatDate(key.expiresAt)}</td>
                      <td className="py-2 pr-4">{key.usageCount}</td>
                      <td className="py-2 pr-4">
                        {formatDateTime(key.lastUsedAt)}
                      </td>
                      <td className="py-2 pr-4">
                        <div className="flex gap-2">
                          <Button
                            variant="secondary"
                            onClick={() => {
                              setTokenModalMode("edit");
                              setEditingTokenId(key.keyId);
                              setTokenForm({
                                ownerEmail: key.ownerEmail,
                                label: key.label,
                                description: key.description,
                                isAdmin: key.isAdmin,
                                rateLimit: key.rateLimit,
                                windowSeconds: key.windowSeconds,
                                expiresAt: key.expiresAt.slice(0, 10),
                              });
                              setTokenModalOpen(true);
                            }}
                          >
                            Edit token
                          </Button>
                          <Button
                            variant="secondary"
                            onClick={async () => {
                              await adminRegenerateKey(key.keyId);
                              showToast("Key regenerated");
                              await loadData();
                            }}
                          >
                            Regenerate
                          </Button>
                          <Button
                            variant="danger"
                            onClick={async () => {
                              await adminRevokeKey(key.keyId);
                              showToast("Key revoked");
                              await loadData();
                            }}
                          >
                            Revoke
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      )}

      {activeTab === "token-management" && (
        <div className="space-y-4">
          <Card>
            <div className="mb-3 flex flex-col gap-2 md:flex-row md:items-center md:justify-between">
              <CardTitle>Usage Overview</CardTitle>
              <div className="w-full md:w-72">
                <Select
                  value={usageUserFilter}
                  onChange={(e) => setUsageUserFilter(e.target.value)}
                >
                  <option value="all">All users</option>
                  {Object.values(usersById)
                    .sort((a, b) => a.email.localeCompare(b.email))
                    .map((user) => (
                      <option key={user.uid} value={user.uid}>
                        {user.displayName} ({user.email})
                      </option>
                    ))}
                </Select>
              </div>
            </div>
          </Card>

          <Card>
            <CardTitle>Daily Usage Stats</CardTitle>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-sm">
                <thead className="border-b border-slate-200 text-slate-500">
                  <tr>
                    <th className="py-2 pr-4">User</th>
                    <th className="py-2 pr-4">Date</th>
                    <th className="py-2 pr-4">Total Requests</th>
                    <th className="py-2 pr-4">Success</th>
                    <th className="py-2 pr-4">Errors</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredUsageStats.map((stat) => (
                    <tr key={stat.statId} className="border-b border-slate-100">
                      <td className="py-2 pr-4">
                        {usersById[stat.userId]?.email || stat.userId}
                      </td>
                      <td className="py-2 pr-4">{stat.date}</td>
                      <td className="py-2 pr-4">{stat.totalRequests}</td>
                      <td className="py-2 pr-4">{stat.successCount}</td>
                      <td className="py-2 pr-4">{stat.errorCount}</td>
                    </tr>
                  ))}
                  {filteredUsageStats.length === 0 && (
                    <tr>
                      <td className="py-3 text-sm text-slate-500" colSpan={5}>
                        No usage stats for this filter.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </Card>

          <Card>
            <CardTitle>Recent Requests</CardTitle>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-sm">
                <thead className="border-b border-slate-200 text-slate-500">
                  <tr>
                    <th className="py-2 pr-4">User</th>
                    <th className="py-2 pr-4">Date/Time</th>
                    <th className="py-2 pr-4">Endpoint</th>
                    <th className="py-2 pr-4">Method</th>
                    <th className="py-2 pr-4">Status Code</th>
                    <th className="py-2 pr-4">Latency ms</th>
                  </tr>
                </thead>
                <tbody>
                  {filteredUsageRequests.map((event) => (
                    <tr key={event.id} className="border-b border-slate-100">
                      <td className="py-2 pr-4">
                        {usersById[event.userId]?.email || event.userId}
                      </td>
                      <td className="py-2 pr-4">
                        {formatDateTime(event.dateTime)}
                      </td>
                      <td className="py-2 pr-4 font-mono text-xs">
                        {event.endpoint}
                      </td>
                      <td className="py-2 pr-4">{event.method}</td>
                      <td className="py-2 pr-4">{event.statusCode}</td>
                      <td className="py-2 pr-4">{event.latencyMs}</td>
                    </tr>
                  ))}
                  {filteredUsageRequests.length === 0 && (
                    <tr>
                      <td className="py-3 text-sm text-slate-500" colSpan={6}>
                        No recent requests for this filter.
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      )}

      {activeTab === "server-controls" && (
        <div className="space-y-4">
          <Card>
            <CardTitle>Server state</CardTitle>
            <div className="mt-3 grid gap-3 text-sm md:grid-cols-2">
              <p>
                <span className="font-medium">Instance ID:</span>{" "}
                {instanceState.instanceId}
              </p>
              <p>
                <span className="font-medium">Type:</span>{" "}
                {instanceState.instanceType}
              </p>
              <p className="flex items-center gap-2">
                <span className="font-medium">State:</span>{" "}
                <StatusBadge status={instanceState.state} />
              </p>
              <p>
                <span className="font-medium">Uptime:</span>{" "}
                {formatUptime(instanceState.uptimeSeconds)}
              </p>
              <p>
                <span className="font-medium">Public IP:</span>{" "}
                {instanceState.publicIp}
              </p>
            </div>
            <div className="mt-4 flex flex-wrap gap-2">
              <Button
                onClick={async () => {
                  await startServer();
                  showToast("Server start requested");
                  await loadData();
                }}
              >
                Start Instance
              </Button>
              <Button
                variant="danger"
                onClick={async () => {
                  await stopServer();
                  showToast("Server stop requested");
                  await loadData();
                }}
              >
                Stop Instance
              </Button>
              <Button
                variant="secondary"
                onClick={async () => {
                  await enableHackutdMode();
                  showToast("HackUTD mode enabled");
                  await loadData();
                }}
              >
                Enable HackUTD mode
              </Button>
              <Button
                variant="secondary"
                onClick={async () => {
                  await disableHackutdMode();
                  showToast("HackUTD mode disabled");
                  await loadData();
                }}
              >
                Disable HackUTD mode
              </Button>
            </div>
          </Card>

          <Card>
            <CardTitle>Server action timeline</CardTitle>
            <div className="mt-3 space-y-2">
              {cronLogs.map((entry, index) => (
                <div
                  key={`${entry.runAt}_${index}`}
                  className="rounded-md border border-slate-200 p-3 text-sm"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-medium">
                      {entry.action.toUpperCase()}
                    </span>
                    <span className="text-slate-500">
                      {formatDateTime(entry.runAt)}
                    </span>
                  </div>
                  <p className="mt-1 text-slate-600">
                    Checked {entry.tokensChecked} tokens, valid{" "}
                    {entry.validTokenCount}, instance {entry.instanceState}
                  </p>
                  {entry.notes && (
                    <p className="mt-1 text-xs text-slate-500">{entry.notes}</p>
                  )}
                </div>
              ))}
            </div>
          </Card>
        </div>
      )}

      {activeTab === "logs-metrics" && (
        <div className="space-y-4">
          <Card>
            <CardTitle>Scraper pipeline logs</CardTitle>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-sm">
                <thead className="border-b border-slate-200 text-slate-500">
                  <tr>
                    <th className="py-2 pr-4">Semester</th>
                    <th className="py-2 pr-4">Status</th>
                    <th className="py-2 pr-4">Courses</th>
                    <th className="py-2 pr-4">Sections</th>
                    <th className="py-2 pr-4">MatchRate</th>
                    <th className="py-2 pr-4">StartedAt</th>
                    <th className="py-2 pr-4">CompletedAt</th>
                    <th className="py-2 pr-4">TriggeredBy</th>
                  </tr>
                </thead>
                <tbody>
                  {scraperLogs.map((log) => (
                    <tr key={log.logId} className="border-b border-slate-100">
                      <td className="py-2 pr-4">{log.semester}</td>
                      <td className="py-2 pr-4">
                        <StatusBadge status={log.status} />
                      </td>
                      <td className="py-2 pr-4">{log.coursesWritten}</td>
                      <td className="py-2 pr-4">{log.sectionsWritten}</td>
                      <td className="py-2 pr-4">
                        {formatPercent(log.mergeMatchRate)}
                      </td>
                      <td className="py-2 pr-4">
                        {formatDateTime(log.startedAt)}
                      </td>
                      <td className="py-2 pr-4">
                        {formatDateTime(log.completedAt)}
                      </td>
                      <td className="py-2 pr-4">{log.triggeredBy}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>

          <Card>
            <CardTitle>Cron log feed</CardTitle>
            <div className="mt-3 space-y-2">
              {cronLogs.map((entry, idx) => (
                <div
                  key={`${entry.runAt}_${idx}`}
                  className="rounded-md border border-slate-200 p-3 text-sm"
                >
                  <p className="font-medium">
                    {entry.date} - {entry.action.toUpperCase()} (
                    {entry.instanceState})
                  </p>
                  <p className="text-slate-600">
                    tokensChecked: {entry.tokensChecked}, validTokenCount:{" "}
                    {entry.validTokenCount}
                  </p>
                  {entry.notes && (
                    <p className="text-xs text-slate-500">{entry.notes}</p>
                  )}
                </div>
              ))}
            </div>
          </Card>

          <Card>
            <CardTitle>Promotion logs</CardTitle>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-sm">
                <thead className="border-b border-slate-200 text-slate-500">
                  <tr>
                    <th className="py-2 pr-4">PromotedAt</th>
                    <th className="py-2 pr-4">DocumentCount</th>
                    <th className="py-2 pr-4">From to To</th>
                    <th className="py-2 pr-4">PromotedBy</th>
                    <th className="py-2 pr-4">Notes</th>
                  </tr>
                </thead>
                <tbody>
                  {promotionLogs.map((log) => (
                    <tr key={log.logId} className="border-b border-slate-100">
                      <td className="py-2 pr-4">
                        {formatDateTime(log.promotedAt)}
                      </td>
                      <td className="py-2 pr-4">{log.documentCount}</td>
                      <td className="py-2 pr-4">{`${log.fromEnv} -> ${log.toEnv}`}</td>
                      <td className="py-2 pr-4">{log.promotedBy}</td>
                      <td className="py-2 pr-4">{log.notes || "-"}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      )}

      {activeTab === "config" && (
        <Card>
          <CardTitle>App Config</CardTitle>
          <div className="mt-4 grid gap-4 md:grid-cols-2">
            <div>
              <Label>Auto-approve mode</Label>
              <Select
                value={configDraft.autoApproveMode}
                onChange={(e) =>
                  setConfigDraft((prev) =>
                    prev
                      ? {
                          ...prev,
                          autoApproveMode: e.target
                            .value as AppConfig["autoApproveMode"],
                        }
                      : prev,
                  )
                }
              >
                <option value="none">None</option>
                <option value="utdallas">UTD only (@utdallas.edu)</option>
                <option value="acmutd">ACM UTD only (@acmutd.co)</option>
                <option value="all">All users</option>
              </Select>
            </div>

            <div>
              <Label>HackUTD mode</Label>
              <Select
                value={configDraft.hackutdModeEnabled ? "on" : "off"}
                onChange={(e) =>
                  setConfigDraft((prev) =>
                    prev
                      ? { ...prev, hackutdModeEnabled: e.target.value === "on" }
                      : prev,
                  )
                }
              >
                <option value="off">Off</option>
                <option value="on">On</option>
              </Select>
            </div>

            <div>
              <Label>HackUTD end date</Label>
              <Input
                type="date"
                value={configDraft.hackutdEndDate.slice(0, 10)}
                onChange={(e) =>
                  setConfigDraft((prev) =>
                    prev
                      ? {
                          ...prev,
                          hackutdEndDate: `${e.target.value}T23:59:59.000Z`,
                        }
                      : prev,
                  )
                }
              />
            </div>

            <div>
              <Label>Current semester</Label>
              <Input
                value={configDraft.currentSemester}
                onChange={(e) =>
                  setConfigDraft((prev) =>
                    prev ? { ...prev, currentSemester: e.target.value } : prev,
                  )
                }
              />
            </div>

            <div>
              <Label>Instance type</Label>
              <Select
                value={configDraft.instanceType}
                onChange={(e) =>
                  setConfigDraft((prev) =>
                    prev
                      ? {
                          ...prev,
                          instanceType: e.target
                            .value as AppConfig["instanceType"],
                        }
                      : prev,
                  )
                }
              >
                <option value="t3.micro">t3.micro</option>
                <option value="t3.small">t3.small</option>
                <option value="t3.large">t3.large</option>
              </Select>
            </div>
          </div>

          <div className="mt-5">
            <Button
              onClick={async () => {
                const next = await updateAppConfig(configDraft);
                setAppConfig(next);
                setConfigDraft(next);
                showToast("Configuration saved");
                await loadData();
              }}
            >
              Save
            </Button>
          </div>
        </Card>
      )}

      <Modal
        open={userModerationModalOpen && !!moderationTarget}
        title={
          moderationTarget?.action === "unban"
            ? "Confirm unban user"
            : "Confirm ban user"
        }
        onClose={() => {
          setUserModerationModalOpen(false);
          setModerationTarget(null);
        }}
      >
        {moderationTarget && (
          <div className="space-y-4">
            <p className="text-sm text-slate-700">
              {moderationTarget.action === "unban"
                ? `Are you sure you want to unban ${moderationTarget.email}?`
                : `Are you sure you want to ban ${moderationTarget.email}? This will set all their keys to inactive.`}
            </p>
            <div className="flex justify-end gap-2">
              <Button
                variant="secondary"
                onClick={() => {
                  setUserModerationModalOpen(false);
                  setModerationTarget(null);
                }}
              >
                Cancel
              </Button>
              <Button
                variant={
                  moderationTarget.action === "unban" ? "secondary" : "danger"
                }
                onClick={async () => {
                  if (moderationTarget.action === "unban") {
                    await unbanUser(moderationTarget.uid);
                    showToast(`User ${moderationTarget.email} unbanned`);
                  } else {
                    await banUser(moderationTarget.uid);
                    showToast(`User ${moderationTarget.email} banned`);
                  }
                  setUserModerationModalOpen(false);
                  setModerationTarget(null);
                  await loadData();
                }}
              >
                Confirm
              </Button>
            </div>
          </div>
        )}
      </Modal>

      <Modal
        open={tokenModalOpen}
        title={tokenModalMode === "add" ? "Add token" : "Edit token"}
        onClose={() => setTokenModalOpen(false)}
      >
        <div className="space-y-3">
          <div>
            <Label>Owner email</Label>
            <Input
              value={tokenForm.ownerEmail}
              disabled={tokenModalMode === "edit"}
              title={
                tokenModalMode === "edit"
                  ? "Owner email cannot be changed when editing a token"
                  : undefined
              }
              onChange={(e) =>
                setTokenForm((prev) => ({
                  ...prev,
                  ownerEmail: e.target.value,
                }))
              }
            />
          </div>
          <div>
            <Label>Label</Label>
            <Input
              value={tokenForm.label}
              onChange={(e) =>
                setTokenForm((prev) => ({ ...prev, label: e.target.value }))
              }
            />
          </div>
          <div>
            <Label>Description</Label>
            <TextArea
              value={tokenForm.description}
              onChange={(e) =>
                setTokenForm((prev) => ({
                  ...prev,
                  description: e.target.value,
                }))
              }
            />
          </div>
          <div className="grid gap-3 md:grid-cols-2">
            <label className="flex items-center gap-2 text-sm">
              <input
                type="checkbox"
                checked={tokenForm.isAdmin}
                onChange={(e) =>
                  setTokenForm((prev) => ({
                    ...prev,
                    isAdmin: e.target.checked,
                  }))
                }
              />
              Is Admin
            </label>
            <div>
              <Label>Rate limit</Label>
              <Input
                type="number"
                min={1}
                value={tokenForm.rateLimit}
                onChange={(e) =>
                  setTokenForm((prev) => ({
                    ...prev,
                    rateLimit: Number(e.target.value) || 1,
                  }))
                }
              />
            </div>
            <div>
              <Label>Window seconds</Label>
              <Input
                type="number"
                min={1}
                value={tokenForm.windowSeconds}
                onChange={(e) =>
                  setTokenForm((prev) => ({
                    ...prev,
                    windowSeconds: Number(e.target.value) || 1,
                  }))
                }
              />
            </div>
            <div>
              <Label>Expiry date</Label>
              <Input
                type="date"
                value={tokenForm.expiresAt}
                onChange={(e) =>
                  setTokenForm((prev) => ({
                    ...prev,
                    expiresAt: e.target.value,
                  }))
                }
              />
            </div>
          </div>

          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => setTokenModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={async () => {
                if (tokenModalMode === "add") {
                  await addToken({
                    ownerEmail: tokenForm.ownerEmail,
                    label: tokenForm.label,
                    description: tokenForm.description,
                    isAdmin: tokenForm.isAdmin,
                    rateLimit: tokenForm.rateLimit,
                    windowSeconds: tokenForm.windowSeconds,
                    expiresAt: `${tokenForm.expiresAt}T23:59:59.000Z`,
                  });
                  showToast("Token added");
                } else if (editingTokenId) {
                  await editToken(editingTokenId, {
                    label: tokenForm.label,
                    description: tokenForm.description,
                    isAdmin: tokenForm.isAdmin,
                    rateLimit: tokenForm.rateLimit,
                    windowSeconds: tokenForm.windowSeconds,
                    expiresAt: `${tokenForm.expiresAt}T23:59:59.000Z`,
                  });
                  showToast("Token updated");
                }
                setTokenModalOpen(false);
                await loadData();
              }}
            >
              Save
            </Button>
          </div>
        </div>
      </Modal>
    </PortalLayout>
  );
}
