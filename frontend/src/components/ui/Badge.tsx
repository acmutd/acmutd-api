import { APIKeyStatus } from "../../types/models";

export function StatusBadge({
  status,
}: {
  status: APIKeyStatus | "running" | "stopped" | "success" | "error";
}) {
  const classes: Record<string, string> = {
    active: "bg-emerald-100 text-emerald-700",
    inactive: "bg-slate-200 text-slate-700",
    pending: "bg-amber-100 text-amber-700",
    rejected: "bg-red-100 text-red-700",
    running: "bg-emerald-100 text-emerald-700",
    stopped: "bg-slate-200 text-slate-700",
    success: "bg-emerald-100 text-emerald-700",
    error: "bg-red-100 text-red-700",
  };

  return (
    <span
      className={`inline-flex rounded-full px-2.5 py-1 text-xs font-semibold uppercase tracking-wide ${classes[status] ?? "bg-slate-100 text-slate-700"}`}
    >
      {status}
    </span>
  );
}
