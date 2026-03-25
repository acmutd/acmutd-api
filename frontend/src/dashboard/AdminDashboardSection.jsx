function formatDate(value) {
  if (!value) {
    return "-";
  }
  return new Date(value).toLocaleString();
}

export function AdminDashboardSection({ requests, keys, approve, reject, deactivate }) {
  return (
    <>
      <section className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <h2 className="text-xl font-semibold text-slate-900">Requests</h2>
        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full border-collapse text-left text-sm">
            <thead className="bg-slate-100 text-xs uppercase tracking-wide text-slate-600">
              <tr>
                <th className="px-3 py-2">UID</th>
                <th className="px-3 py-2">Email</th>
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
                  <td className="px-3 py-2 text-slate-700">{formatDate(item.requested_at)}</td>
                  <td className="px-3 py-2 text-slate-700">{item.request_type || "-"}</td>
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
        <h2 className="text-xl font-semibold text-slate-900">Issued Keys</h2>
        <div className="mt-4 overflow-x-auto">
          <table className="min-w-full border-collapse text-left text-sm">
            <thead className="bg-slate-100 text-xs uppercase tracking-wide text-slate-600">
              <tr>
                <th className="px-3 py-2">UID</th>
                <th className="px-3 py-2">Email</th>
                <th className="px-3 py-2">Masked Key</th>
                <th className="px-3 py-2">Status</th>
                <th className="px-3 py-2">Usage</th>
                <th className="px-3 py-2">Updated</th>
                <th className="px-3 py-2">Action</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-200">
              {keys.map((item) => (
                <tr key={item.uid}>
                  <td className="px-3 py-2 font-mono text-xs text-slate-800">{item.uid || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.email || "-"}</td>
                  <td className="px-3 py-2 font-mono text-xs text-slate-800">{item.masked_key || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{item.status || "-"}</td>
                  <td className="px-3 py-2 text-slate-700">{String(item.usage_count || 0)}</td>
                  <td className="px-3 py-2 text-slate-700">{formatDate(item.updated_at)}</td>
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
    </>
  );
}
