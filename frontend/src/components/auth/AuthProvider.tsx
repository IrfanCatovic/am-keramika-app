"use client";

import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
} from "react";

import { apiRequest, ApiError } from "@/lib/api";
import { clearToken, getToken, setToken } from "@/lib/auth-token";
import {
  AuthUser,
  isUserRole,
  LoginResponse,
  UserRole,
} from "@/types/auth";

interface AuthContextValue {
  user: AuthUser | null;
  token: string | null;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
}

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

function normalizeUser(payload: AuthUser): AuthUser {
  if (!isUserRole(payload.role)) {
    throw new ApiError("Nepoznata uloga korisnika.", 500, payload);
  }
  return {
    id: payload.id,
    username: payload.username,
    role: payload.role as UserRole,
    isActive: payload.isActive,
  };
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<AuthUser | null>(null);
  const [token, setTokenState] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const logout = useCallback(() => {
    clearToken();
    setTokenState(null);
    setUser(null);
  }, []);

  useEffect(() => {
    let cancelled = false;

    async function restoreSession() {
      const existing = getToken();
      if (!existing) {
        if (!cancelled) {
          setIsLoading(false);
        }
        return;
      }

      if (!cancelled) {
        setTokenState(existing);
      }

      try {
        const me = await apiRequest<AuthUser>("/auth/me");
        if (!cancelled) {
          setUser(normalizeUser(me));
        }
      } catch {
        clearToken();
        if (!cancelled) {
          setTokenState(null);
          setUser(null);
        }
      } finally {
        if (!cancelled) {
          setIsLoading(false);
        }
      }
    }

    // Odgodi bootstrap da se setState ne poziva sinhrono u effect tijelu.
    const timer = window.setTimeout(() => {
      void restoreSession();
    }, 0);

    return () => {
      cancelled = true;
      window.clearTimeout(timer);
    };
  }, []);

  const login = useCallback(async (username: string, password: string) => {
    const response = await apiRequest<LoginResponse>("/auth/login", {
      method: "POST",
      auth: false,
      body: { username, password },
    });

    setToken(response.token);
    setTokenState(response.token);
    setUser(normalizeUser(response.user));
  }, []);

  const value = useMemo(
    () => ({
      user,
      token,
      isLoading,
      login,
      logout,
    }),
    [user, token, isLoading, login, logout],
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextValue {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error("useAuth mora biti korišten unutar AuthProvider.");
  }
  return context;
}
