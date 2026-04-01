import { Link, useNavigate } from "react-router-dom";
import { AppConfig, User } from "../../types/models";
import { Button } from "../ui/Button";
import { useEffect, useRef, useState } from "react";

interface NavItem {
  key: string;
  label: string;
}

export function PortalLayout({
  user,
  appConfig,
  title,
  navItems,
  activeTab,
  onTabChange,
  onSignOut,
  children,
}: {
  user: User;
  appConfig: AppConfig;
  title: string;
  navItems: NavItem[];
  activeTab: string;
  onTabChange: (tab: string) => void;
  onSignOut: () => Promise<void>;
  children: React.ReactNode;
}) {
  const [menuOpen, setMenuOpen] = useState(false);
  const menuContainerRef = useRef<HTMLDivElement | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    if (!menuOpen) {
      return;
    }

    function handlePointerDown(event: MouseEvent) {
      if (!menuContainerRef.current) {
        return;
      }
      const target = event.target as Node;
      if (!menuContainerRef.current.contains(target)) {
        setMenuOpen(false);
      }
    }

    document.addEventListener("mousedown", handlePointerDown);
    return () => {
      document.removeEventListener("mousedown", handlePointerDown);
    };
  }, [menuOpen]);

  return (
    <div className="flex h-screen bg-slate-50 text-slate-900">
      <aside className="hidden w-64 flex-col border-r border-slate-200 bg-white md:flex">
        <div className="border-b border-slate-200 p-4 text-lg font-bold">
          ACM UTD API
        </div>
        <nav className="flex-1 space-y-1 p-3">
          {navItems.map((item) => (
            <button
              key={item.key}
              onClick={() => onTabChange(item.key)}
              className={`w-full rounded-md px-3 py-2 text-left text-sm font-medium transition ${
                activeTab === item.key
                  ? "bg-slate-900 text-white"
                  : "text-slate-700 hover:bg-slate-100"
              }`}
            >
              {item.label}
            </button>
          ))}
        </nav>
      </aside>

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-16 items-center justify-between border-b border-slate-200 bg-white px-4 md:px-6">
          <div>
            <Link to="/" className="text-base font-bold text-slate-900">
              ACM UTD API
            </Link>
            <p className="text-xs text-slate-500">{title}</p>
          </div>

          <div className="hidden items-center gap-2 md:flex">
            <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-700">
              Semester: {appConfig.currentSemester}
            </span>
            <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-medium text-slate-700">
              {appConfig.instanceType}
            </span>
            {appConfig.hackutdModeEnabled && (
              <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-medium text-amber-700">
                HackUTD
              </span>
            )}
          </div>

          <div className="relative" ref={menuContainerRef}>
            <button
              className="flex items-center gap-2 rounded-md border border-slate-200 bg-white px-2 py-1.5"
              onClick={() => setMenuOpen((v) => !v)}
            >
              <div className="hidden text-left md:block">
                <p className="text-sm font-medium leading-4">
                  {user.displayName}
                </p>
                <p className="text-xs text-slate-500">{user.email}</p>
              </div>
            </button>

            {menuOpen && (
              <div className="absolute right-0 top-0 z-30 w-52 rounded-md border border-slate-200 bg-white p-2 shadow-md">
                <div className="mb-1 rounded px-3 py-2 text-left">
                  <p className="flex items-center justify-between text-sm font-medium text-slate-900">
                    <span>{user.displayName}</span>
                    <span className="text-xs text-slate-500">
                      {user.isAdmin ? "Admin" : "User"}
                    </span>
                  </p>
                  <p className="text-xs text-slate-500">{user.email}</p>

                  <p className="text-xs text-slate-500">
                    Status: {user.approvalStatus}
                  </p>
                </div>
                <button
                  className="w-full rounded px-3 py-2 text-left text-sm hover:bg-slate-100"
                  onClick={() => {
                    navigate("/");
                    void onSignOut();
                  }}
                >
                  Sign out
                </button>
              </div>
            )}
          </div>
        </header>

        <div className="md:hidden border-b border-slate-200 bg-white p-2">
          <div className="flex gap-2 overflow-x-auto">
            {navItems.map((item) => (
              <Button
                key={item.key}
                variant={activeTab === item.key ? "primary" : "secondary"}
                className="whitespace-nowrap"
                onClick={() => onTabChange(item.key)}
              >
                {item.label}
              </Button>
            ))}
          </div>
        </div>

        <main className="min-h-0 flex-1 overflow-auto p-4 md:p-6">
          {children}
        </main>
      </div>
    </div>
  );
}
