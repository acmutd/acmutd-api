export function UserDashboardSection({
  hasActiveKey,
  registrationMode,
  keyInfo,
  generatedKey,
  loading,
  requestOrRegenerateKey,
  revokeKey
}) {
  return (
    <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
      <div className="flex flex-wrap items-center gap-3">
        <h2 className="text-xl font-semibold text-slate-900">Key Overview</h2>
        <span
          className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${
            hasActiveKey
              ? "bg-emerald-100 text-emerald-700"
              : "bg-slate-200 text-slate-700"
          }`}
        >
          {hasActiveKey ? "active" : "inactive"}
        </span>
      </div>

      <p className="mt-2 text-sm text-slate-600">Registration mode: {registrationMode}</p>

      {keyInfo && (
        <dl className="mt-4 grid gap-3 rounded-xl border border-slate-200 bg-slate-50 p-4 sm:grid-cols-2">
          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Masked API key</dt>
            <dd className="mt-1 font-mono text-sm text-slate-900">{keyInfo.masked_key || "-"}</dd>
          </div>

          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Rate limit</dt>
            <dd className="mt-1 text-sm text-slate-900">{String(keyInfo.rate_limit || "-")}</dd>
          </div>

          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Window seconds</dt>
            <dd className="mt-1 text-sm text-slate-900">{String(keyInfo.window_seconds || "-")}</dd>
          </div>

          <div>
            <dt className="text-xs font-semibold uppercase tracking-wide text-slate-500">Usage count</dt>
            <dd className="mt-1 text-sm text-slate-900">{String(keyInfo.usage_count || 0)}</dd>
          </div>
        </dl>
      )}

      {generatedKey && (
        <div className="mt-4 space-y-2">
          <p className="text-sm text-slate-600">New key generated. This is shown once:</p>
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
        {keyInfo && (
          <button
            className="rounded-lg bg-rose-600 px-4 py-2 text-sm font-semibold text-white transition hover:bg-rose-700 disabled:cursor-not-allowed disabled:opacity-60"
            onClick={revokeKey}
            disabled={loading}
          >
            Revoke key
          </button>
        )}
      </div>
    </section>
  );
}
