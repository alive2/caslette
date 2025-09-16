import React, { createContext, useContext, useEffect, useState } from "react";
import type { User } from "../types";
import { authApi } from "../services/api";

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (username: string, password: string) => Promise<void>;
  logout: () => void;
  loading: boolean;
  isAdmin: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error("useAuth must be used within an AuthProvider");
  }
  return context;
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [user, setUser] = useState<User | null>(null);
  const [token, setToken] = useState<string | null>(
    localStorage.getItem("admin_token")
  );
  const [loading, setLoading] = useState(true);

  const isAdmin =
    user?.roles?.some(
      (role) => role.name === "admin" || role.name === "moderator"
    ) ?? false;

  useEffect(() => {
    const initAuth = async () => {
      if (token) {
        try {
          const profileData = await authApi.getProfile();
          setUser(profileData.user);
        } catch {
          localStorage.removeItem("admin_token");
          setToken(null);
        }
      }
      setLoading(false);
    };

    initAuth();
  }, [token]);

  const login = async (username: string, password: string) => {
    const response = await authApi.login({ username, password });

    // Check if user has admin or moderator role
    const hasAdminRole = response.user.roles?.some(
      (role) => role.name === "admin" || role.name === "moderator"
    );

    if (!hasAdminRole) {
      throw new Error("Access denied: Admin privileges required");
    }

    localStorage.setItem("admin_token", response.token);
    setToken(response.token);
    setUser(response.user);
  };

  const logout = () => {
    localStorage.removeItem("admin_token");
    setToken(null);
    setUser(null);
  };

  const value = {
    user,
    token,
    login,
    logout,
    loading,
    isAdmin,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};
