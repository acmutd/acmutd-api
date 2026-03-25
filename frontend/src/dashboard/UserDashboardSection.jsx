function formatDate(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

function StatusBadge({ active, activeLabel = "active", inactiveLabel = "inactive" }) {
  return (
    <span
      className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${
        active ? "bg-emerald-100 text-emerald-700" : "bg-slate-200 text-slate-700"
      }`}
    >
      {active ? activeLabel : inactiveLabel}
    </span>
  );
}

export function UserDashboardSection({
  user,
  auth,
  hasActiveKey,
  registrationMode,
  keyInfo,
  usage,
  generatedKey,
  expiryNotifications,
  hackutdMode,
  allowedActions,
  loading,
  requestOrRegenerateKey,
  revokeKey
}) {
  const endpointBreakdown = usage?.endpoint_breakdown || [];
  const recentRequests = usage?.recent_requests || [];

  return (
    <>
      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-xl font-semibold text-slate-900">Authentication & Eligibility</h2>
            <p className="mt-1 text-sm text-slate-600">Google OAuth via Firebase Auth with UTD domain enforcement.</p>
          </div>
          <StatusBadge active={auth?.firebase_token_status === "valid"} activeLabel="token valid" inactiveLabel="token invalid" />
        </div>

        <dl className="mt-4 grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Signed in user</dt>
            <dd className="mt-1 text-slate-900">{user?.name || user?.email || "-"}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Email domain policy</dt>
            <dd className="mt-1 text-slate-900">@{auth?.allowed_domain || "-"}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">OAuth provider</dt>
            <dd className="mt-1 text-slate-900">{user?.auth_provider || "google.com"}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Registration mode</dt>
            <dd className="mt-1 text-slate-900">{registrationMode}</dd>
          </div>
        </dl>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-wrap items-center gap-3">
            <h2 className="text-xl font-semibold text-slate-900">Key Management</h2>
            <StatusBadge active={hasActiveKey} />
          </div>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-700">
            Limit: {String(allowedActions?.max_keys_per_user || 1)} key per user
          </span>
        </div>

        <div className="mt-4 grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Masked key</p>
            <p className="mt-1 font-mono text-slate-900">{keyInfo?.masked_key || "-"}</p>
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Created</p>
            <p className="mt-1 text-slate-900">{formatDate(keyInfo?.created_at)}</p>
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Expires</p>
            <p className="mt-1 text-slate-900">{formatDate(keyInfo?.expires_at)}</p>
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Key type</p>
            <p className="mt-1 text-slate-900">{keyInfo?.key_type || "standard"}</p>
          </div>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-2">
          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Request New Key</h3>
            <p className="mt-1 text-xs text-slate-600">Provide identity from OAuth and a short label/description for admin review.</p>
            <div className="mt-3 space-y-3">
              <div>
                <label className="text-xs font-semibold uppercase tracking-wide text-slate-500">Owner identity</label>
                <input
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700"
                  value={user?.email || ""}
                  disabled
                  readOnly
                />
              </div>
              <div>
                <label className="text-xs font-semibold uppercase tracking-wide text-slate-500">Key label</label>
                <input
                  className="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-700"
                  value={keyInfo?.request_description || "Example: ACM bot integration"}
                  disabled
                  readOnly
                />
              </div>
              <p className="text-xs text-slate-500">UI skeleton only: wired for POST /dashboard/key in production.</p>
            </div>
          </div>

          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Lifecycle Rules</h3>
            <ul className="mt-3 space-y-2 text-sm text-slate-700">
              <li>Semester policy: keys expire after the current semester.</li>
              <li>HackUTD mode auto-approves and sets expiry to the day after the event.</li>
              <li>Regeneration is atomic: old key deactivated before new key issuance.</li>
              <li>Revoke removes access immediately through DELETE /dashboard/key.</li>
            </ul>
            <div className="mt-3 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-800">
              HackUTD mode: {hackutdMode?.enabled ? "enabled" : "disabled"} | forced expiry {formatDate(hackutdMode?.forced_expiry_at)}
            </div>
          </div>
        </div>

        {generatedKey && (
          <div className="mt-4 space-y-2">
            <p className="text-sm text-slate-600">New key generated. This full value is shown once:</p>
            <pre className="overflow-x-auto rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 font-mono text-xs text-amber-900">
              {generatedKey}
            </pre>
          </div>
        )}

        <div className="mt-6 flex flex-wrap items-center gap-3">
          <button
            className="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-sky-700 disabled:cursor-not-allowed disabled:opacity-60"
            onClick={requestOrRegenerateKey}
            disabled={loading}
          >
            {hasActiveKey ? "Regenerate key" : "Request API key"}
          </button>
          <button
            className="rounded-lg bg-rose-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-60"
            onClick={revokeKey}
            disabled={loading || !keyInfo}
          >
            Revoke key
          </button>
        </div>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <h2 className="text-xl font-semibold text-slate-900">Expiry Warning Email</h2>
        <div className="mt-4 grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm sm:grid-cols-3">
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Email warnings</p>
            <p className="mt-1 text-slate-900">{expiryNotifications?.email_warning_enabled ? "enabled" : "disabled"}</p>
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Warning windows</p>
            <p className="mt-1 text-slate-900">{(expiryNotifications?.warning_windows_days || []).join(", ")} days</p>
          </div>
          <div>
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Next warning</p>
            <p className="mt-1 text-slate-900">{formatDate(expiryNotifications?.next_warning_at)}</p>
          </div>
        </div>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-xl font-semibold text-slate-900">Usage Statistics (Stretch Goal UI)</h2>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-700">
            Static debug data
          </span>
        </div>

        <div className="mt-4 grid gap-3 sm:grid-cols-3">
          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Total requests</p>
            <p className="mt-2 text-2xl font-bold text-slate-900">{String(usage?.total_requests || 0)}</p>
          </div>
          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Last used</p>
            <p className="mt-2 text-sm font-semibold text-slate-900">{formatDate(usage?.last_used_at)}</p>
          </div>
          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4">
            <p className="text-xs font-semibold uppercase tracking-wide text-slate-500">Top endpoint</p>
            <p className="mt-2 text-sm font-semibold text-slate-900">
              {endpointBreakdown.length ? endpointBreakdown[0].endpoint : "-"}
            </p>
          </div>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-2">
          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Most Frequent Endpoints</h3>
            <ul className="mt-3 space-y-2">
              {endpointBreakdown.map((entry) => (
                <li key={entry.endpoint}>
                  <div className="flex items-center justify-between text-xs text-slate-600">
                    <span className="font-mono">{entry.endpoint}</span>
                    <span>{entry.requests} req</span>
                  </div>
                  <div className="mt-1 h-2 rounded-full bg-slate-200">
                    <div
                      className="h-2 rounded-full bg-sky-500"
                      style={{ width: `${Math.min(100, Math.max(8, (entry.requests / (usage?.total_requests || 1)) * 100))}%` }}
                    />
                  </div>
                </li>
              ))}
            </ul>
          </div>

          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Recent Requests</h3>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-xs">
                <thead className="text-slate-500">
                  <tr>
                    <th className="py-1 pr-3">Timestamp</th>
                    <th className="py-1 pr-3">Endpoint</th>
                    <th className="py-1 pr-1">Status</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {recentRequests.map((request, index) => (
                    <tr key={`${request.endpoint}-${index}`}>
                      <td className="py-2 pr-3 text-slate-700">{formatDate(request.at)}</td>
                      <td className="py-2 pr-3 font-mono text-slate-700">{request.endpoint}</td>
                      <td className="py-2 pr-1">
                        <span
                          className={`rounded-full px-2 py-0.5 font-semibold ${
                            request.status >= 400 ? "bg-rose-100 text-rose-700" : "bg-emerald-100 text-emerald-700"
                          }`}
                        >
                          {String(request.status)}
                        </span>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>
    </>
  );
}
