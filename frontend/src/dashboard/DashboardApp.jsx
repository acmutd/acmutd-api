import { useState } from "react";
import { UserDashboardSection } from "./UserDashboardSection";
import { AdminDashboardSection } from "./AdminDashboardSection";
import { debugUserDashboardData } from "./debug/userDashboardData";
import { debugAdminDashboardData } from "./debug/adminDashboardData";

export function DashboardApp() {
  const [debugViewRole, setDebugViewRole] = useState("user");
  const isAdminView = debugViewRole === "admin";
  const activeUser = isAdminView ? debugAdminDashboardData.user : debugUserDashboardData.user;
  const userDashboard = debugUserDashboardData.dashboard;
  const userAuth = debugUserDashboardData.auth;
  const allowedActions = debugUserDashboardData.allowed_actions;

  const status = userDashboard?.status || "none";
  const keyInfo = userDashboard?.key_info || null;
  const hasActiveKey = status === "active";
  const registrationMode = userDashboard?.mode || "request-only";

  function switchDebugView() {
    setDebugViewRole((previous) => (previous === "admin" ? "user" : "admin"));
  }

  function noop() {}

  return (
    <div className="mx-auto max-w-7xl px-6 py-10 sm:px-8">
      <main className="rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">ACM UTD Dashboard Skeleton</h1>
          <span
            className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${
              isAdminView ? "bg-emerald-100 text-emerald-700" : "bg-slate-200 text-slate-700"
            }`}
          >
            {isAdminView ? "admin" : "user"}
          </span>
        </div>
        <p className="mt-2 text-sm text-slate-600 sm:text-base">
          Debug-only UI preview for Cloudflare Pages frontend. All sections use static fixtures under src/dashboard/debug.
        </p>

        <div className="mt-4 flex flex-wrap items-center gap-3 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2">
          <span className="text-xs font-semibold uppercase tracking-wide text-amber-700">Debug mode</span>
          <button
            className="rounded-lg border border-amber-300 bg-white px-3 py-1.5 text-xs font-semibold text-amber-700 transition hover:bg-amber-100"
            onClick={switchDebugView}
          >
            Switch to {isAdminView ? "user" : "admin"} view
          </button>
        </div>

        <div className="mt-5 flex flex-wrap items-center gap-3">
          {activeUser && (
            <span
              className="rounded-lg border border-slate-300 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700"
            >
              Signed in as {activeUser.email || ""}
            </span>
          )}
          <span className="rounded-lg border border-slate-300 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700">
            Frontend auth target: Firebase ID token + Google OAuth (@utdallas.edu)
          </span>
        </div>
      </main>

      {!isAdminView && (
        <UserDashboardSection
          user={debugUserDashboardData.user}
          auth={userAuth}
          hasActiveKey={hasActiveKey}
          registrationMode={registrationMode}
          keyInfo={keyInfo}
          usage={userDashboard?.usage}
          generatedKey={debugUserDashboardData.generatedKey}
          expiryNotifications={userDashboard?.expiry_notifications}
          hackutdMode={userDashboard?.hackutd_mode}
          allowedActions={allowedActions}
          loading={false}
          requestOrRegenerateKey={noop}
          revokeKey={noop}
        />
      )}

      {isAdminView && (
        <AdminDashboardSection
          user={debugAdminDashboardData.user}
          auth={debugAdminDashboardData.auth}
          approval={debugAdminDashboardData.approval}
          requests={debugAdminDashboardData.requests}
          keys={debugAdminDashboardData.keys}
          manualTokenForm={debugAdminDashboardData.manual_token_form}
          server={debugAdminDashboardData.server}
          scraperPipeline={debugAdminDashboardData.scraperPipeline}
          analytics={debugAdminDashboardData.analytics}
          scraperLogs={debugAdminDashboardData.scraperLogs}
          cronLogs={debugAdminDashboardData.cronLogs}
          approve={noop}
          reject={noop}
          deactivate={noop}
        />
      )}
    </div>
  );
}
