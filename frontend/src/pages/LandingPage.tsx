import { useEffect } from "react";
import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { useAuth } from "../state/AuthContext";

export function LandingPage() {
  const { login, currentUser, loading, switchToAdmin, switchToStudent } =
    useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    if (!loading && currentUser) {
      navigate("/dashboard");
    }
  }, [currentUser, loading, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
      <div className="w-full max-w-xl rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
          API Portal
        </p>
        <h1 className="mt-2 text-3xl font-bold text-slate-900">ACM UTD API</h1>
        <p className="mt-2 text-slate-600">Courses and grades data for UTD.</p>

        <div className="mt-8 grid gap-3">
          <Button
            onClick={async () => {
              await login();
              navigate("/dashboard");
            }}
          >
            Sign in with Google
          </Button>
        </div>
      </div>
    </div>
  );
}
