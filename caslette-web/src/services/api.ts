import axios from "axios";
import type { AuthResponse, LoginRequest, RegisterRequest } from "../types";

const API_BASE_URL = "http://localhost:8080/api/v1";

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Add token to requests if available
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
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

  register: async (data: RegisterRequest): Promise<AuthResponse> => {
    const response = await api.post("/auth/register", data);
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
};

export const diamondApi = {
  getUserDiamonds: async (userId: number) => {
    const response = await api.get(`/diamonds/user/${userId}`);
    return response.data;
  },
};

export default api;
