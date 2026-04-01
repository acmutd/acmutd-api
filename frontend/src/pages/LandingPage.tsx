import { useNavigate } from "react-router-dom";
import { Button } from "../components/ui/Button";
import { useAuth } from "../state/AuthContext";
import { useToast } from "../state/ToastContext";

export function LandingPage() {
  const { login, switchToAdmin, switchToStudent } = useAuth();
  const navigate = useNavigate();
  const { showToast } = useToast();

  return (
    <div className="flex min-h-screen items-center justify-center bg-slate-50 px-4">
      <div className="w-full max-w-xl rounded-2xl border border-slate-200 bg-white p-8 shadow-sm">
        <p className="text-xs font-semibold uppercase tracking-[0.18em] text-slate-500">
          API Portal
        </p>
        <h1 className="mt-2 text-3xl font-bold text-slate-900">ACM UTD API</h1>
        <p className="mt-2 text-slate-600">
          Course and grade data for UTD projects and HackUTD.
        </p>

        <div className="mt-8 grid gap-3">
          <Button
            onClick={async () => {
              await login();
              showToast("Signed in with mocked Google OAuth");
              navigate("/dashboard");
            }}
          >
            Sign in with Google
          </Button>
        </div>

        <div className="mt-6 border-t border-slate-200 pt-4">
          <p className="mb-2 text-xs text-slate-500">
            Developer switch (mock users)
          </p>
          <div className="flex gap-2">
            <Button
              variant="ghost"
              onClick={async () => {
                await switchToStudent();
                navigate("/dashboard");
              }}
            >
              Use Student
            </Button>
            <Button
              variant="ghost"
              onClick={async () => {
                await switchToAdmin();
                navigate("/dashboard");
              }}
            >
              Use Admin
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
