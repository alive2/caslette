import axios from "axios";
import type {
  AuthResponse,
  LoginRequest,
  DiamondTransactionRequest,
} from "../types";

const API_BASE_URL = "http://localhost:8080/api/v1";

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Add token to requests if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("admin_token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export const authApi = {
  login: async (data: LoginRequest): Promise<AuthResponse> => {
    const response = await api.post("/auth/login", data);
    return response.data;
  },

  getProfile: async () => {
    const response = await api.get("/auth/profile");
    return response.data;
  },
};

export const userApi = {
  getUsers: async (page = 1, limit = 10) => {
    const response = await api.get(`/users?page=${page}&limit=${limit}`);
    return response.data;
  },

  getUser: async (id: number) => {
    const response = await api.get(`/users/${id}`);
    return response.data;
  },

  updateUser: async (id: number, data: Record<string, unknown>) => {
    const response = await api.put(`/users/${id}`, data);
    return response.data;
  },

  deleteUser: async (id: number) => {
    const response = await api.delete(`/users/${id}`);
    return response.data;
  },

  assignRoles: async (id: number, roleIds: number[]) => {
    const response = await api.post(`/users/${id}/roles`, {
      role_ids: roleIds,
    });
    return response.data;
  },

  removeRole: async (userId: number, roleId: number) => {
    const response = await api.delete(`/users/${userId}/roles/${roleId}`);
    return response.data;
  },

  assignPermissions: async (id: number, permissionIds: number[]) => {
    const response = await api.post(`/users/${id}/permissions`, {
      permission_ids: permissionIds,
    });
    return response.data;
  },

  getUserPermissions: async (id: number) => {
    const response = await api.get(`/users/${id}/permissions`);
    return response.data;
  },

  removePermission: async (userId: number, permissionId: number) => {
    const response = await api.delete(
      `/users/${userId}/permissions/${permissionId}`
    );
    return response.data;
  },
};

export const roleApi = {
  getRoles: async () => {
    const response = await api.get("/roles");
    return response.data;
  },

  getRole: async (id: number) => {
    const response = await api.get(`/roles/${id}`);
    return response.data;
  },

  createRole: async (data: { name: string; description: string }) => {
    const response = await api.post("/roles", data);
    return response.data;
  },

  updateRole: async (
    id: number,
    data: { name?: string; description?: string }
  ) => {
    const response = await api.put(`/roles/${id}`, data);
    return response.data;
  },

  deleteRole: async (id: number) => {
    const response = await api.delete(`/roles/${id}`);
    return response.data;
  },

  assignPermissions: async (roleId: number, permissionIds: number[]) => {
    const response = await api.post(`/roles/${roleId}/permissions`, {
      permission_ids: permissionIds,
    });
    return response.data;
  },

  removePermission: async (roleId: number, permissionId: number) => {
    const response = await api.delete(
      `/roles/${roleId}/permissions/${permissionId}`
    );
    return response.data;
  },
};

export const permissionApi = {
  getPermissions: async () => {
    const response = await api.get("/permissions");
    return response.data;
  },

  getPermission: async (id: number) => {
    const response = await api.get(`/permissions/${id}`);
    return response.data;
  },
};

export const diamondApi = {
  getUserDiamonds: async (userId: number) => {
    const response = await api.get(`/diamonds/user/${userId}`);
    return response.data;
  },

  getAllTransactions: async (page = 1, limit = 50) => {
    const response = await api.get(
      `/diamonds/transactions?page=${page}&limit=${limit}`
    );
    return response.data;
  },

  creditDiamonds: async (data: DiamondTransactionRequest) => {
    const response = await api.post("/diamonds/credit", data);
    return response.data;
  },

  debitDiamonds: async (data: DiamondTransactionRequest) => {
    const response = await api.post("/diamonds/debit", data);
    return response.data;
  },
};

export default api;
