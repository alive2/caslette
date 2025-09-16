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
