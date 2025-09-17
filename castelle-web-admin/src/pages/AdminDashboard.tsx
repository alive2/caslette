import React, { useEffect, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { userApi, diamondApi } from "../services/api";
import type { User, Diamond, DiamondTransactionRequest } from "../types";

const AdminDashboard: React.FC = () => {
  const { user, logout } = useAuth();
  const [users, setUsers] = useState<User[]>([]);
  const [diamonds, setDiamonds] = useState<Diamond[]>([]);
  const [loading, setLoading] = useState(true);
  const [activeTab, setActiveTab] = useState<"users" | "diamonds">("users");

  // Diamond transaction form
  const [diamondForm, setDiamondForm] = useState({
    user_id: 0,
    amount: 0,
    type: "credit" as "credit" | "debit",
    description: "",
  });

  useEffect(() => {
    fetchUsers();
    fetchDiamonds();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await userApi.getUsers(1, 50);
      setUsers(response.users);
    } catch {
      console.error("Failed to fetch users");
    } finally {
      setLoading(false);
    }
  };

  const fetchDiamonds = async () => {
    try {
      const response = await diamondApi.getAllTransactions(1, 50);
      setDiamonds(response.transactions);
    } catch {
      console.error("Failed to fetch diamonds");
    }
  };

  const handleDiamondTransaction = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const request: DiamondTransactionRequest = {
        user_id: diamondForm.user_id,
        amount: Math.abs(diamondForm.amount),
        type: diamondForm.type,
        description: diamondForm.description,
      };

      if (diamondForm.type === "credit") {
        await diamondApi.creditDiamonds(request);
      } else {
        await diamondApi.debitDiamonds(request);
      }

      setDiamondForm({
        user_id: 0,
        amount: 0,
        type: "credit",
        description: "",
      });
      fetchDiamonds();
      alert("Transaction completed successfully!");
    } catch (error) {
      alert("Transaction failed. Please try again.");
    }
  };

  const handleToggleUserStatus = async (
    userId: number,
    currentStatus: boolean
  ) => {
    try {
      await userApi.updateUser(userId, { is_active: !currentStatus });
      fetchUsers();
    } catch (error) {
      alert("Failed to update user status");
    }
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        Loading...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* Header */}
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-gray-900">
                Caslette Admin
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <span className="text-sm text-gray-500">
                Welcome, {user?.username}
              </span>
              <button
                onClick={logout}
                className="bg-red-600 hover:bg-red-700 text-white px-4 py-2 rounded-md text-sm font-medium"
              >
                Logout
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Navigation Tabs */}
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 mt-6">
        <div className="border-b border-gray-200">
          <nav className="-mb-px flex space-x-8">
            <button
              onClick={() => setActiveTab("users")}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === "users"
                  ? "border-indigo-500 text-indigo-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              User Management
            </button>
            <button
              onClick={() => setActiveTab("diamonds")}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === "diamonds"
                  ? "border-indigo-500 text-indigo-600"
                  : "border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300"
              }`}
            >
              Diamond Management
            </button>
          </nav>
        </div>
      </div>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6">
        {activeTab === "users" && (
          <div className="space-y-6">
            {/* Users Table */}
            <div className="bg-white shadow overflow-hidden sm:rounded-md">
              <div className="px-4 py-5 sm:px-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900">
                  Users
                </h3>
                <p className="mt-1 max-w-2xl text-sm text-gray-500">
                  Manage user accounts and their status
                </p>
              </div>
              <ul className="divide-y divide-gray-200">
                {users.map((user) => (
                  <li key={user.id} className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        <div className="flex-shrink-0">
                          <div className="w-10 h-10 bg-indigo-100 rounded-full flex items-center justify-center">
                            <span className="text-indigo-600 font-medium text-sm">
                              {user.username.charAt(0).toUpperCase()}
                            </span>
                          </div>
                        </div>
                        <div className="ml-4">
                          <div className="text-sm font-medium text-gray-900">
                            {user.username}
                          </div>
                          <div className="text-sm text-gray-500">
                            {user.email}
                          </div>
                          <div className="text-xs text-gray-400">
                            Roles:{" "}
                            {user.roles?.map((r) => r.name).join(", ") ||
                              "None"}
                          </div>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            user.is_active
                              ? "bg-green-100 text-green-800"
                              : "bg-red-100 text-red-800"
                          }`}
                        >
                          {user.is_active ? "Active" : "Inactive"}
                        </span>
                        <button
                          onClick={() =>
                            handleToggleUserStatus(user.id, user.is_active)
                          }
                          className={`text-sm px-3 py-1 rounded ${
                            user.is_active
                              ? "bg-red-100 text-red-700 hover:bg-red-200"
                              : "bg-green-100 text-green-700 hover:bg-green-200"
                          }`}
                        >
                          {user.is_active ? "Deactivate" : "Activate"}
                        </button>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}

        {activeTab === "diamonds" && (
          <div className="space-y-6">
            {/* Diamond Transaction Form */}
            <div className="bg-white shadow sm:rounded-lg">
              <div className="px-4 py-5 sm:p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Diamond Transaction
                </h3>
                <form
                  onSubmit={handleDiamondTransaction}
                  className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4"
                >
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      User
                    </label>
                    <select
                      value={diamondForm.user_id}
                      onChange={(e) =>
                        setDiamondForm({
                          ...diamondForm,
                          user_id: parseInt(e.target.value),
                        })
                      }
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                      required
                    >
                      <option value="">Select user</option>
                      {users.map((user) => (
                        <option key={user.id} value={user.id}>
                          {user.username}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Type
                    </label>
                    <select
                      value={diamondForm.type}
                      onChange={(e) =>
                        setDiamondForm({
                          ...diamondForm,
                          type: e.target.value as "credit" | "debit",
                        })
                      }
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                    >
                      <option value="credit">Credit</option>
                      <option value="debit">Debit</option>
                    </select>
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700">
                      Amount
                    </label>
                    <input
                      type="number"
                      min="1"
                      value={diamondForm.amount}
                      onChange={(e) =>
                        setDiamondForm({
                          ...diamondForm,
                          amount: parseInt(e.target.value),
                        })
                      }
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                      required
                    />
                  </div>
                  <div className="sm:col-span-2 lg:col-span-1 flex items-end">
                    <button
                      type="submit"
                      className="w-full bg-indigo-600 text-white py-2 px-4 rounded-md hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500"
                    >
                      Execute
                    </button>
                  </div>
                  <div className="sm:col-span-2">
                    <label className="block text-sm font-medium text-gray-700">
                      Description
                    </label>
                    <input
                      type="text"
                      value={diamondForm.description}
                      onChange={(e) =>
                        setDiamondForm({
                          ...diamondForm,
                          description: e.target.value,
                        })
                      }
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-indigo-500 focus:border-indigo-500"
                      placeholder="Optional description"
                    />
                  </div>
                </form>
              </div>
            </div>

            {/* Recent Transactions */}
            <div className="bg-white shadow overflow-hidden sm:rounded-md">
              <div className="px-4 py-5 sm:px-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900">
                  Recent Transactions
                </h3>
              </div>
              <ul className="divide-y divide-gray-200">
                {diamonds.slice(0, 20).map((diamond) => (
                  <li key={diamond.id} className="px-4 py-4 sm:px-6">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center">
                        <div className="flex-shrink-0">
                          <span
                            className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                              diamond.amount > 0
                                ? "bg-green-100 text-green-800"
                                : "bg-red-100 text-red-800"
                            }`}
                          >
                            {diamond.type}
                          </span>
                        </div>
                        <div className="ml-4">
                          <div className="text-sm font-medium text-gray-900">
                            User ID: {diamond.user_id}
                          </div>
                          <div className="text-sm text-gray-500">
                            {diamond.description || "No description"}
                          </div>
                          <div className="text-xs text-gray-400">
                            {new Date(diamond.created_at).toLocaleString()}
                          </div>
                        </div>
                      </div>
                      <div className="text-right">
                        <div
                          className={`text-sm font-medium ${
                            diamond.amount > 0
                              ? "text-green-600"
                              : "text-red-600"
                          }`}
                        >
                          {diamond.amount > 0 ? "+" : ""}
                          {diamond.amount.toLocaleString()}
                        </div>
                        <div className="text-sm text-gray-500">
                          Balance: {diamond.balance.toLocaleString()}
                        </div>
                      </div>
                    </div>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default AdminDashboard;
