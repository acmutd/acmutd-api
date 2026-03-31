import { Link } from "react-router-dom";

export function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 p-6">
      <div className="rounded-xl border border-slate-200 bg-white p-8 text-center">
        <h1 className="text-2xl font-bold text-slate-900">Page not found</h1>
        <p className="mt-2 text-sm text-slate-600">
          This route is not part of the API portal skeleton.
        </p>
        <Link
          to="/"
          className="mt-4 inline-block text-sm font-medium text-slate-900 underline"
        >
          Back to landing
        </Link>
      </div>
    </div>
  );
}
