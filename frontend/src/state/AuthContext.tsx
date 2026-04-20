import {
  GoogleAuthProvider,
  onAuthStateChanged,
  signInWithPopup,
} from "firebase/auth";
import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import { flushSync } from "react-dom";
import { firebaseAuth } from "../lib/firebase";
import { setSessionUser } from "../lib/apiClient";
import { User } from "../types/models";

interface AuthContextValue {
  currentUser: User | null;
  loading: boolean;
  login: () => Promise<void>;
  logout: () => Promise<void>;
  switchToAdmin: () => Promise<void>;
  switchToStudent: () => Promise<void>;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [currentUser, setCurrentUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const unsubscribe = onAuthStateChanged(firebaseAuth, async (fbUser) => {
      if (fbUser) {
        try {
          const token = await fbUser.getIdToken();
          const res = await fetch("/dashboard/api/v1/users/me", {
            headers: { Authorization: `Bearer ${token}` },
          });
          setCurrentUser(res.ok ? ((await res.json()) as User) : null);
        } catch {
          setCurrentUser(null);
        }
      } else {
        setCurrentUser(null);
      }
      setLoading(false);
    });
    return unsubscribe;
  }, []);

  const login = useCallback(async () => {
    const result = await signInWithPopup(firebaseAuth, new GoogleAuthProvider());
    const token = await result.user.getIdToken();
    const res = await fetch("/dashboard/api/v1/users/me", {
      headers: { Authorization: `Bearer ${token}` },
    });
    if (res.ok) {
      const userData = (await res.json()) as User;
      // flushSync commits the state update synchronously so RequireAuth
      // sees the new user before navigate() triggers its render.
      flushSync(() => setCurrentUser(userData));
    }
  }, []);

  const logout = useCallback(async () => {
    await firebaseAuth.signOut();
  }, []);

  // Dev-only helpers for switching between mock users.
  // These are no-ops when apiClient.ts points to the real backend.
  const switchToAdmin = useCallback(async () => {
    if (!import.meta.env.DEV) return;
    const user = await setSessionUser("u_002").catch(() => null);
    if (user) setCurrentUser(user);
  }, []);

  const switchToStudent = useCallback(async () => {
    if (!import.meta.env.DEV) return;
    const user = await setSessionUser("u_001").catch(() => null);
    if (user) setCurrentUser(user);
  }, []);

  const value = useMemo(
    () => ({
      currentUser,
      loading,
      login,
      logout,
      switchToAdmin,
      switchToStudent,
    }),
    [currentUser, loading, login, logout, switchToAdmin, switchToStudent]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) {
    throw new Error("useAuth must be used within AuthProvider");
  }
  return ctx;
}
