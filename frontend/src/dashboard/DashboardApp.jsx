import { useState } from "react";
import { UserDashboardSection } from "./UserDashboardSection";
import { AdminDashboardSection } from "./AdminDashboardSection";
import { debugUserDashboardData } from "./debug/userDashboardData";
import { debugAdminDashboardData } from "./debug/adminDashboardData";

export function DashboardApp() {
  const [debugViewRole, setDebugViewRole] = useState("user");
  const isAdminView = debugViewRole === "admin";
  const userDashboard = debugUserDashboardData.dashboard;
  const userAuth = debugUserDashboardData.auth;
  const allowedActions = debugUserDashboardData.allowed_actions;

  // Admin can view user dashboard for their own account + admin-specific features
  const activeUser = isAdminView ? debugAdminDashboardData.user : debugUserDashboardData.user;
  const adminAsUser = isAdminView ? debugAdminDashboardData.user : null;

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
          <h1 className="text-3xl font-bold tracking-tight text-slate-900">ACM API Dashboard Skeleton</h1>
          <div className="flex flex-wrap items-center gap-3">
            <a
              href="/"
              className="rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-semibold text-slate-700 transition hover:bg-slate-100"
            >
              Go to Documentation
            </a>
            <span
              className={`inline-flex rounded-full px-3 py-1 text-xs font-semibold uppercase tracking-wide ${
                isAdminView ? "bg-emerald-100 text-emerald-700" : "bg-slate-200 text-slate-700"
              }`}
            >
              {isAdminView ? "admin" : "user"}
            </span>
          </div>
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
              Signed in as {activeUser.email || ""} {isAdminView && "(admin account)"}
            </span>
          )}
          <span className="rounded-lg border border-slate-300 bg-slate-50 px-4 py-2 text-sm font-semibold text-slate-700">
            Frontend auth target: Firebase ID token + Google OAuth (open sign-up; key domain groups for approvals)
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
        <>
          <div className="mt-6 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
            <h2 className="text-2xl font-bold tracking-tight text-slate-900">Your User Dashboard</h2>
            <p className="mt-1 text-sm text-slate-600">View and manage your own keys as a user would, even while having admin privileges.</p>
          </div>
          <UserDashboardSection
            user={adminAsUser}
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

          <div className="mt-10 rounded-2xl border border-slate-200 bg-white/90 p-7 shadow-sm backdrop-blur">
            <h2 className="text-2xl font-bold tracking-tight text-slate-900">Admin Controls & Monitoring</h2>
            <p className="mt-1 text-sm text-slate-600">Officer-only operations for key approvals, server management, and analytics.</p>
          </div>

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
        </>
      )}
    </div>
  );
}
