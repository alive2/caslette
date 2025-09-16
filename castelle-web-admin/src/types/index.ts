export interface User {
  id: number;
  username: string;
  email: string;
  first_name: string;
  last_name: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  roles: Role[];
}

export interface Role {
  id: number;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
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
