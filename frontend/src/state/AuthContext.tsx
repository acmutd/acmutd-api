import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";
import {
  getCurrentUser,
  loginWithGoogle,
  setSessionUser,
  signOut,
} from "../lib/api";
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
    async function bootstrap() {
      const user = await getCurrentUser();
      setCurrentUser(user);
      setLoading(false);
    }
    void bootstrap();
  }, []);

  const login = useCallback(async () => {
    const user = await loginWithGoogle();
    setCurrentUser(user);
  }, []);

  const logout = useCallback(async () => {
    await signOut();
    setCurrentUser(null);
  }, []);

  const switchToAdmin = useCallback(async () => {
    const user = await setSessionUser("u_002");
    setCurrentUser(user);
  }, []);

  const switchToStudent = useCallback(async () => {
    const user = await setSessionUser("u_001");
    setCurrentUser(user);
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
    [currentUser, loading, login, logout, switchToAdmin, switchToStudent],
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
