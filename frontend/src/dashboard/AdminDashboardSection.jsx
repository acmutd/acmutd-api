function formatDate(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

function miniBarWidth(current, maxValue) {
  if (!maxValue) {
    return "8%";
  }
  return `${Math.min(100, Math.max(8, (current / maxValue) * 100))}%`;
}

export function AdminDashboardSection({
  user,
  auth,
  approval,
  requests,
  keys,
  manualTokenForm,
  server,
  scraperPipeline,
  analytics,
  scraperLogs,
  cronLogs,
  approve,
  reject,
  deactivate
}) {
  const dailySeries = analytics?.requests_per_day || [];
  const byToken = analytics?.by_token || [];
  const byEndpoint = analytics?.by_endpoint || [];
  const dayMax = dailySeries.reduce((maxValue, day) => Math.max(maxValue, day.count), 0);
  const tokenMax = byToken.reduce((maxValue, item) => Math.max(maxValue, item.requests), 0);
  const endpointMax = byEndpoint.reduce((maxValue, item) => Math.max(maxValue, item.requests), 0);

  return (
    <>
      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-xl font-semibold text-slate-900">Admin Access & Policy</h2>
            <p className="mt-1 text-sm text-slate-600">Officer access is gated by Firestore isAdmin with Google OAuth. Account creation is open to all domains.</p>
          </div>
          <span className="inline-flex rounded-full bg-emerald-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-emerald-700">
            {auth?.admin_gate_status || "granted"}
          </span>
        </div>

        <dl className="mt-4 grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 text-sm sm:grid-cols-2 lg:grid-cols-4">
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Signed in admin</dt>
            <dd className="mt-1 text-slate-900">{user?.name || user?.email || "-"}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Provider</dt>
            <dd className="mt-1 text-slate-900">{auth?.oauth_provider || "google.com"}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Account domain policy</dt>
            <dd className="mt-1 text-slate-900">{auth?.domain_restriction === "none" ? "Open sign-up (all domains)" : `@${auth?.domain_restriction}`}</dd>
          </div>
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Approval mode</dt>
            <dd className="mt-1 text-slate-900">{approval?.mode || "manual"}</dd>
          </div>
        </dl>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-xl font-semibold text-slate-900">Approval Queue</h2>
          <div className="flex items-center gap-2 rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
            <span className="text-xs font-semibold uppercase tracking-wide text-slate-600">Manual approval</span>
            <button className="rounded border border-slate-300 bg-white px-2 py-1 text-xs font-semibold text-slate-700 transition hover:bg-slate-100">
              Toggle to {approval?.mode === "manual" ? "auto-approve" : "manual"}
            </button>
          </div>
        </div>

        <p className="mt-2 text-sm text-slate-600">{approval?.global_toggle_hint}</p>

        <div className="mt-4 rounded-xl border border-slate-200 bg-slate-50 p-4">
          <h3 className="text-sm font-semibold text-slate-900">Domain Auto-Approve Rules</h3>
          <p className="mt-1 text-xs text-slate-600">Configure which request domains bypass manual review.</p>
          <div className="mt-3 grid gap-2 sm:grid-cols-3">
            <button className="flex items-center justify-between rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-semibold text-slate-700">
              <span>@acmutd.co</span>
              <span className={`rounded-full px-2 py-0.5 ${approval?.auto_approve_by_domain?.acmutd_co ? "bg-emerald-100 text-emerald-700" : "bg-slate-200 text-slate-700"}`}>
                {approval?.auto_approve_by_domain?.acmutd_co ? "auto" : "manual"}
              </span>
            </button>
            <button className="flex items-center justify-between rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-semibold text-slate-700">
              <span>@utdallas.edu</span>
              <span className={`rounded-full px-2 py-0.5 ${approval?.auto_approve_by_domain?.utdallas_edu ? "bg-emerald-100 text-emerald-700" : "bg-slate-200 text-slate-700"}`}>
                {approval?.auto_approve_by_domain?.utdallas_edu ? "auto" : "manual"}
              </span>
            </button>
            <button className="flex items-center justify-between rounded-lg border border-slate-300 bg-white px-3 py-2 text-xs font-semibold text-slate-700">
              <span>All other domains</span>
              <span className={`rounded-full px-2 py-0.5 ${approval?.auto_approve_by_domain?.other_domains ? "bg-emerald-100 text-emerald-700" : "bg-slate-200 text-slate-700"}`}>
                {approval?.auto_approve_by_domain?.other_domains ? "auto" : "manual"}
              </span>
            </button>
          </div>
        </div>

        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full border-collapse text-left text-sm">
            <thead className="bg-slate-100 text-xs uppercase tracking-wide text-slate-600">
              <tr>
                <th className="px-3 py-2">UID</th>
                <th className="px-3 py-2">Email</th>
                <th className="px-3 py-2">Label</th>
                <th className="px-3 py-2">Description</th>
                <th className="px-3 py-2">Requested</th>
                <th className="px-3 py-2">Type</th>
                <th className="px-3 py-2">Status</th>
                <th className="px-3 py-2">Actions</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {requests.map((item) => (
                <tr key={item.uid}>
                  <td className="px-3 py-2 font-mono text-xs text-slate-800">{item.uid || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.email || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.label || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.description || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{formatDate(item.requested_at)}</td>
                  <td className="px-3 py-2 text-slate-700">{item.requested_key_type || item.request_type || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.status || "-"}</td>
                  <td className="px-3 py-2">
                    <div className="flex flex-wrap gap-2">
                      <button
                        className="rounded bg-emerald-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-emerald-700"
                        onClick={() => approve(item.uid)}
                      >
                        Approve
                      </button>
                      <button
                        className="rounded bg-rose-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-rose-700"
                        onClick={() => reject(item.uid)}
                      >
                        Reject
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-xl font-semibold text-slate-900">Token Management</h2>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-700">
            Admins may hold multiple keys
          </span>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-3">
          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Add Token Manually</h3>
            <p className="mt-1 text-xs text-slate-600">Create random token or provide custom token value.</p>
            <div className="mt-3 space-y-2 text-xs text-slate-700">
              <div className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Label: {manualTokenForm?.label_placeholder}</div>
              <div className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Type: {(manualTokenForm?.key_type_options || []).join(" / ")}</div>
              <div className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Expiry presets: {(manualTokenForm?.expiry_presets || []).join(" | ")}</div>
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <button className="rounded bg-sky-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-sky-700">Generate random</button>
              <button className="rounded border border-slate-300 bg-white px-3 py-1.5 text-xs font-semibold text-slate-700 transition hover:bg-slate-100">Use custom value</button>
            </div>
          </div>

          <div className="rounded-xl border border-slate-200 p-4 lg:col-span-2">
            <h3 className="text-sm font-semibold text-slate-900">Token Edit Actions</h3>
            <ul className="mt-3 grid gap-2 text-sm text-slate-700 sm:grid-cols-2">
              <li className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Update label and description</li>
              <li className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Adjust expiry date and semester policy</li>
              <li className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Toggle active/inactive state</li>
              <li className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">Deactivate or remove token permanently</li>
            </ul>
          </div>
        </div>

        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full border-collapse text-left text-sm">
            <thead className="bg-slate-100 text-xs uppercase tracking-wide text-slate-600">
              <tr>
                <th className="px-3 py-2">UID</th>
                <th className="px-3 py-2">Email</th>
                <th className="px-3 py-2">Label</th>
                <th className="px-3 py-2">Type</th>
                <th className="px-3 py-2">Requested</th>
                <th className="px-3 py-2">Expiry</th>
                <th className="px-3 py-2">Status</th>
                <th className="px-3 py-2">Usage</th>
                <th className="px-3 py-2">Last used</th>
                <th className="px-3 py-2">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {keys.map((item) => (
                <tr key={item.uid}>
                  <td className="px-3 py-2 font-mono text-xs text-slate-800">{item.uid || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.email || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.label || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.type || "standard"}</td>
                  <td className="px-3 py-2 text-slate-700">{formatDate(item.created_at)}</td>
                  <td className="px-3 py-2 text-slate-700">{formatDate(item.expires_at)}</td>
                  <td className="px-3 py-2 text-slate-700">{item.status || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{String(item.usage_count || 0)}</td>
                  <td className="px-3 py-2 text-slate-700">{formatDate(item.last_used_at)}</td>
                  <td className="px-3 py-2">
                    <button
                      className="rounded bg-rose-600 px-3 py-1.5 text-xs font-semibold text-white transition hover:bg-rose-700"
                      onClick={() => deactivate(item.uid)}
                    >
                      Deactivate
                    </button>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-xl font-semibold text-slate-900">Server Controls</h2>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-700">
            AWS SDK-backed backend routes
          </span>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-3">
          <div className="rounded-xl border border-slate-200 bg-slate-50 p-4 lg:col-span-2">
            <h3 className="text-sm font-semibold text-slate-900">Server State Panel</h3>
            <dl className="mt-3 grid gap-2 text-sm sm:grid-cols-2">
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Instance id</dt>
                <dd className="mt-1 font-mono text-slate-900">{server?.ec2_instance_id || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">State</dt>
                <dd className="mt-1 text-slate-900">{server?.state || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Instance type</dt>
                <dd className="mt-1 text-slate-900">{server?.instance_type || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Public IP</dt>
                <dd className="mt-1 font-mono text-slate-900">{server?.public_ip || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Uptime</dt>
                <dd className="mt-1 text-slate-900">{server?.uptime_human || "-"}</dd>
              </div>
              <div>
                <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">HackUTD mode</dt>
                <dd className="mt-1 text-slate-900">{server?.hackutd_mode_enabled ? "enabled" : "disabled"}</dd>
              </div>
            </dl>
          </div>

          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Actions</h3>
            <div className="mt-3 grid gap-2">
              <button className="rounded bg-emerald-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-emerald-700">Start server</button>
              <button className="rounded bg-rose-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-rose-700">Stop server</button>
              <button className="rounded border border-slate-300 bg-white px-3 py-2 text-xs font-semibold text-slate-700 transition hover:bg-slate-100">Enable HackUTD mode</button>
              <button className="rounded border border-slate-300 bg-white px-3 py-2 text-xs font-semibold text-slate-700 transition hover:bg-slate-100">Disable HackUTD mode</button>
              <button className="rounded bg-sky-600 px-3 py-2 text-xs font-semibold text-white transition hover:bg-sky-700">Run scraper + pipeline</button>
            </div>
          </div>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-2">
          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">HackUTD Enable Sequence</h3>
            <ol className="mt-3 list-decimal space-y-1 pl-4 text-sm text-slate-700">
              {(server?.hackutd_actions || []).map((step) => (
                <li key={`hackutd-${step}`}>{step}</li>
              ))}
            </ol>
          </div>
          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">HackUTD Disable Sequence</h3>
            <ol className="mt-3 list-decimal space-y-1 pl-4 text-sm text-slate-700">
              {(server?.normal_mode_actions || []).map((step) => (
                <li key={`normal-${step}`}>{step}</li>
              ))}
            </ol>
          </div>
        </div>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h2 className="text-xl font-semibold text-slate-900">Statistics & Monitoring</h2>
          <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-700">
            Static charts from debug fixtures
          </span>
        </div>

        <div className="mt-4 grid gap-4 lg:grid-cols-2">
          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Requests Per Day + Error Count</h3>
            <ul className="mt-3 space-y-2">
              {dailySeries.map((day) => (
                <li key={day.day}>
                  <div className="flex items-center justify-between text-xs text-slate-600">
                    <span>{day.day}</span>
                    <span>{day.count} req / {day.errors} errors</span>
                  </div>
                  <div className="mt-1 h-2 rounded-full bg-slate-200">
                    <div className="h-2 rounded-full bg-sky-500" style={{ width: miniBarWidth(day.count, dayMax) }} />
                  </div>
                </li>
              ))}
            </ul>
            <p className="mt-3 text-xs text-slate-600">Error rate over time (current): {String(analytics?.error_rate_percent || 0)}%</p>
          </div>

          <div className="rounded-xl border border-slate-200 p-4">
            <h3 className="text-sm font-semibold text-slate-900">Requests By Token</h3>
            <ul className="mt-3 space-y-2">
              {byToken.map((entry) => (
                <li key={entry.label}>
                  <div className="flex items-center justify-between text-xs text-slate-600">
                    <span>{entry.label}</span>
                    <span>{entry.requests}</span>
                  </div>
                  <div className="mt-1 h-2 rounded-full bg-slate-200">
                    <div className="h-2 rounded-full bg-emerald-500" style={{ width: miniBarWidth(entry.requests, tokenMax) }} />
                  </div>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="mt-4 rounded-xl border border-slate-200 p-4">
          <h3 className="text-sm font-semibold text-slate-900">Requests By Endpoint</h3>
          <ul className="mt-3 grid gap-2 sm:grid-cols-2">
            {byEndpoint.map((entry) => (
              <li key={entry.endpoint} className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2">
                <div className="flex items-center justify-between text-xs text-slate-600">
                  <span className="font-mono text-slate-700">{entry.endpoint}</span>
                  <span>{entry.requests}</span>
                </div>
                <div className="mt-1 h-2 rounded-full bg-slate-200">
                  <div className="h-2 rounded-full bg-indigo-500" style={{ width: miniBarWidth(entry.requests, endpointMax) }} />
                </div>
              </li>
            ))}
          </ul>
        </div>
      </section>

      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="grid gap-4 lg:grid-cols-2">
          <div className="rounded-xl border border-slate-200 p-4">
            <h2 className="text-lg font-semibold text-slate-900">Scraper/Pipeline Log Feed</h2>
            <p className="mt-1 text-xs text-slate-600">
              Last run {formatDate(scraperPipeline?.last_run_at)} | Uploaded {String(scraperPipeline?.courses_uploaded || 0)} courses | Match rate {scraperPipeline?.merge_match_rate}
            </p>
            <div className="mt-3 max-h-56 space-y-2 overflow-y-auto rounded-lg bg-slate-950 p-3 font-mono text-xs text-slate-100">
              {(scraperLogs || []).map((entry, index) => (
                <div key={`${entry.time}-${index}`}>
                  <span className="text-slate-400">[{formatDate(entry.time)}]</span> {entry.level.toUpperCase()} {entry.message}
                </div>
              ))}
            </div>
          </div>

          <div className="rounded-xl border border-slate-200 p-4">
            <h2 className="text-lg font-semibold text-slate-900">Lambda Cron Log Feed</h2>
            <p className="mt-1 text-xs text-slate-600">Static preview of cronLogs from Firestore.</p>
            <div className="mt-3 overflow-x-auto">
              <table className="min-w-full text-left text-xs">
                <thead className="text-slate-500">
                  <tr>
                    <th className="py-1 pr-2">Date</th>
                    <th className="py-1 pr-2">Checked</th>
                    <th className="py-1 pr-2">Valid</th>
                    <th className="py-1 pr-1">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {(cronLogs || []).map((entry, index) => (
                    <tr key={`${entry.time}-${index}`}>
                      <td className="py-2 pr-2 text-slate-700">{formatDate(entry.time)}</td>
                      <td className="py-2 pr-2 text-slate-700">{String(entry.tokens_checked)}</td>
                      <td className="py-2 pr-2 text-slate-700">{String(entry.valid_count)}</td>
                      <td className="py-2 pr-1 text-slate-700">{entry.action}</td>
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
