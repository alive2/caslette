export interface User {
  id: number;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  is_active: boolean;
  balance: number;
  created_at: string;
  updated_at: string;
  roles: Role[];
  permissions?: Permission[]; // User-specific permissions
}

export interface Role {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  permissions?: Permission[];
}

export interface Permission {
  id: number;
  name: string;
  description: string;
  resource: string;
  action: string;
  created_at: string;
  updated_at: string;
}

export interface RolePermission {
  role_id: number;
  permission_id: number;
  created_at: string;
}

export interface UserPermission {
  user_id: number;
  permission_id: number;
  created_at: string;
}

export interface CreateRoleRequest {
  name: string;
  description: string;
}

export interface UpdateRoleRequest {
  name?: string;
  description?: string;
}

// Default roles in the system
export const DEFAULT_ROLES = {
  PLAYER: "player",
  MANAGER: "manager",
  ADMIN: "admin",
} as const;

export const DEFAULT_ROLE = DEFAULT_ROLES.PLAYER;

export interface Diamond {
  id: number;
  user_id: number;
  amount: number;
  balance: number;
  transaction_id: string;
  type: string;
  description: string;
  created_at: string;
  updated_at: string;
  user?: User;
}

export interface AuthResponse {
  token: string;
  user: User;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface DiamondTransactionRequest {
  user_id: number;
  amount: number;
  type: string;
  description?: string;
}
