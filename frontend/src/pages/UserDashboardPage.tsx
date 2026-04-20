import { useEffect, useMemo, useState } from "react";
import { PortalLayout } from "../components/layout/PortalLayout";
import { StatusBadge } from "../components/ui/Badge";
import { Button } from "../components/ui/Button";
import { Card, CardSubtitle, CardTitle } from "../components/ui/Card";
import { Input, Label, TextArea } from "../components/ui/Field";
import { Modal } from "../components/ui/Modal";
import {
  getAppConfig,
  getMyApiKey,
  getMyDailyStats,
  getMyRecentRequests,
  regenerateApiKey,
  requestApiKey,
  revokeApiKey,
} from "../lib/apiClient";
import { formatDate, formatDateTime, maskKey } from "../lib/format";
import { useAuth } from "../state/AuthContext";
import { useToast } from "../state/ToastContext";
import { APIKey, AppConfig, DailyStat, RequestEvent } from "../types/models";

const userNavItems = [
  { key: "overview", label: "Overview" },
  { key: "my-key", label: "My API Key" },
  { key: "usage", label: "Usage" },
];

export function UserDashboardPage() {
  const { currentUser, logout } = useAuth();
  const { showToast } = useToast();
  const [activeTab, setActiveTab] = useState("overview");
  const [appConfig, setAppConfig] = useState<AppConfig | null>(null);
  const [myKey, setMyKey] = useState<APIKey | null>(null);
  const [stats, setStats] = useState<DailyStat[]>([]);
  const [requests, setRequests] = useState<RequestEvent[]>([]);
  const [loading, setLoading] = useState(true);

  const [requestModalOpen, setRequestModalOpen] = useState(false);
  const [requestLabel, setRequestLabel] = useState("");
  const [requestDescription, setRequestDescription] = useState("");

  const [confirmAction, setConfirmAction] = useState<
    "regenerate" | "revoke" | null
  >(null);

  const [keyRevealed, setKeyRevealed] = useState(false);

  const total30Days = stats.reduce((sum, item) => sum + item.totalRequests, 0);

  const mostUsedEndpoint = useMemo(() => {
    const counts = new Map<string, number>();
    stats.forEach((s) => {
      s.endpointBreakdown.forEach((e) => {
        counts.set(e.endpoint, (counts.get(e.endpoint) ?? 0) + e.count);
      });
    });
    let best = "N/A";
    let max = -1;
    counts.forEach((count, endpoint) => {
      if (count > max) {
        max = count;
        best = endpoint;
      }
    });
    return best;
  }, [stats]);

  useEffect(() => {
    if (!currentUser) {
      return;
    }
    const user = currentUser;

    async function load() {
      setLoading(true);
      const [cfg, key, daily, recent] = await Promise.all([
        getAppConfig(),
        getMyApiKey(user.uid),
        getMyDailyStats(user.uid),
        getMyRecentRequests(user.uid),
      ]);
      setAppConfig(cfg);
      setMyKey(key);
      setStats(daily);
      setRequests(recent);
      setKeyRevealed(false);
      setLoading(false);
    }

    void load();
  }, [currentUser]);

  if (!currentUser) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <p className="text-sm text-slate-600">
          Please sign in from landing page.
        </p>
      </div>
    );
  }

  if (loading || !appConfig) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-slate-50">
        <p className="text-sm text-slate-600">Loading dashboard...</p>
      </div>
    );
  }

  return (
    <PortalLayout
      user={currentUser}
      appConfig={appConfig}
      title="User Dashboard"
      navItems={userNavItems}
      activeTab={activeTab}
      onTabChange={setActiveTab}
      onSignOut={logout}
    >
      {activeTab === "overview" && (
        <div className="flex h-full flex-col">
          <Card>
            <CardTitle>Overview</CardTitle>
            <div>This will have general info</div>
          </Card>
        </div>
      )}

      {activeTab === "my-key" && (
        <div className="space-y-4">
          {!myKey ? (
            <Card>
              <CardTitle>You don't have an API key yet</CardTitle>
              <CardSubtitle>
                Request access for your project to start using ACM UTD API.
              </CardSubtitle>
              <div className="mt-4">
                <Button onClick={() => setRequestModalOpen(true)}>
                  Request API Key
                </Button>
              </div>
            </Card>
          ) : (
            <Card>
              <div className="mb-4 flex items-center justify-between">
                <CardTitle>{myKey.label}</CardTitle>
                <StatusBadge status={myKey.status} />
              </div>
              <div className="grid gap-3 text-sm md:grid-cols-2">
                <p>
                  <span className="font-medium">Description:</span>{" "}
                  {myKey.description}
                </p>
                <p>
                  <span className="font-medium">Admin Flag:</span>{" "}
                  {myKey.isAdmin ? "Yes" : "No"}
                </p>
                <p>
                  <span className="font-medium">Rate Limit:</span>{" "}
                  {myKey.rateLimit} per {myKey.windowSeconds}s
                </p>
                <p>
                  <span className="font-medium">Lifetime Usage:</span>{" "}
                  {myKey.usageCount.toLocaleString()}
                </p>
                <p>
                  <span className="font-medium">Created At:</span>{" "}
                  {formatDateTime(myKey.createdAt)}
                </p>
                <p>
                  <span className="font-medium">Expires At:</span>{" "}
                  {formatDateTime(myKey.expiresAt)}
                </p>
                <p>
                  <div
                    className="inline-flex items-center gap-1"
                    onMouseLeave={() => setKeyRevealed(false)}
                  >
                    <button
                      type="button"
                      className="break-all font-mono text-sm text-slate-800 underline decoration-dotted underline-offset-2"
                      onClick={() => setKeyRevealed((v) => !v)}
                    >
                      {keyRevealed ? myKey.key : maskKey(myKey.key)}
                    </button>
                    <button
                      type="button"
                      title="Copy to clipboard"
                      className={`text-slate-400 hover:text-slate-700 ${keyRevealed ? "" : "invisible pointer-events-none"}`}
                      onClick={async () => {
                        try {
                          await navigator.clipboard.writeText(myKey.key);
                          showToast("Copied key to clipboard");
                        } catch {
                          showToast("Unable to copy key");
                        }
                      }}
                    >
                      <svg
                        xmlns="http://www.w3.org/2000/svg"
                        className="h-3.5 w-3.5"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                      >
                        <rect
                          x="9"
                          y="9"
                          width="13"
                          height="13"
                          rx="2"
                          ry="2"
                        />
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                      </svg>
                    </button>
                  </div>
                </p>
              </div>

              <div className="mt-5 flex flex-wrap gap-2">
                <Button onClick={() => setConfirmAction("regenerate")}>
                  Regenerate key
                </Button>
                <Button
                  variant="danger"
                  onClick={() => setConfirmAction("revoke")}
                >
                  Delete / Revoke key
                </Button>
              </div>
            </Card>
          )}
        </div>
      )}

      {activeTab === "usage" && (
        <div className="space-y-4">
          <div className="grid gap-4 md:grid-cols-3">
            <Card>
              <CardSubtitle>Total Requests (last 30 days)</CardSubtitle>
              <CardTitle>{total30Days.toLocaleString()}</CardTitle>
            </Card>
            <Card>
              <CardSubtitle>Most Used Endpoint</CardSubtitle>
              <CardTitle>{mostUsedEndpoint}</CardTitle>
            </Card>
            <Card>
              <CardSubtitle>Last Used</CardSubtitle>
              <CardTitle>{formatDateTime(myKey?.lastUsedAt ?? null)}</CardTitle>
            </Card>
          </div>

          <Card>
            <CardTitle>Daily Usage Stats</CardTitle>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-sm">
                <thead className="border-b border-slate-200 text-slate-500">
                  <tr>
                    <th className="py-2 pr-4">Date</th>
                    <th className="py-2 pr-4">Total Requests</th>
                    <th className="py-2 pr-4">Success</th>
                    <th className="py-2 pr-4">Errors</th>
                  </tr>
                </thead>
                <tbody>
                  {stats.map((s) => (
                    <tr key={s.statId} className="border-b border-slate-100">
                      <td className="py-2 pr-4">{s.date}</td>
                      <td className="py-2 pr-4">{s.totalRequests}</td>
                      <td className="py-2 pr-4">{s.successCount}</td>
                      <td className="py-2 pr-4">{s.errorCount}</td>
                    </tr>
                  ))}
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
                    <th className="py-2 pr-4">Date/Time</th>
                    <th className="py-2 pr-4">Endpoint</th>
                    <th className="py-2 pr-4">Method</th>
                    <th className="py-2 pr-4">Status Code</th>
                    <th className="py-2 pr-4">Latency ms</th>
                  </tr>
                </thead>
                <tbody>
                  {requests.map((event) => (
                    <tr key={event.id} className="border-b border-slate-100">
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
                </tbody>
              </table>
            </div>
          </Card>
        </div>
      )}

      <Modal
        open={requestModalOpen}
        title="Request API Key"
        onClose={() => setRequestModalOpen(false)}
      >
        <div className="space-y-3">
          <div>
            <Label>Key label</Label>
            <Input
              value={requestLabel}
              onChange={(e) => setRequestLabel(e.target.value)}
              placeholder="HackUTD Team API"
            />
          </div>
          <div>
            <Label>Description / project name</Label>
            <TextArea
              value={requestDescription}
              onChange={(e) => setRequestDescription(e.target.value)}
              placeholder="Describe your project"
            />
          </div>
          <div className="flex justify-end gap-2">
            <Button
              variant="secondary"
              onClick={() => setRequestModalOpen(false)}
            >
              Cancel
            </Button>
            <Button
              onClick={async () => {
                await requestApiKey(requestLabel, requestDescription);
                const key = await getMyApiKey(currentUser.uid);
                setMyKey(key);
                setRequestModalOpen(false);
                setRequestLabel("");
                setRequestDescription("");
                showToast("Key requested (pending approval)");
              }}
            >
              Submit
            </Button>
          </div>
        </div>
      </Modal>

      <Modal
        open={Boolean(confirmAction)}
        title="Are you sure?"
        onClose={() => {
          setConfirmAction(null);
        }}
      >
        <p className="text-sm text-slate-600">
          {confirmAction === "regenerate"
            ? "This will generate a new token value and invalidate the previous key."
            : "This will revoke and deactivate the key."}
        </p>
        <div className="mt-4 flex justify-end gap-2">
          <Button variant="secondary" onClick={() => setConfirmAction(null)}>
            Cancel
          </Button>
          <Button
            variant={confirmAction === "revoke" ? "danger" : "primary"}
            onClick={async () => {
              if (!myKey || !confirmAction) {
                return;
              }
              if (confirmAction === "regenerate") {
                const updated = await regenerateApiKey(myKey.keyId);
                setMyKey(updated);
                setKeyRevealed(false);
                showToast("API key regenerated");
              }
              if (confirmAction === "revoke") {
                await revokeApiKey(myKey.keyId);
                const next = await getMyApiKey(currentUser.uid);
                setMyKey(next);
                showToast("API key revoked");
              }
              setConfirmAction(null);
            }}
          >
            Confirm
          </Button>
        </div>
      </Modal>
    </PortalLayout>
  );
}
