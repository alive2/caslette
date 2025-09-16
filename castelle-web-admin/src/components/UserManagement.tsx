import React, { useEffect, useState } from "react";
import AdminLayout from "./AdminLayout";
import { userApi, diamondApi } from "../services/api";
import type { User } from "../types";
import { FaGem } from "react-icons/fa";

interface UserManagementState {
  users: User[];
  loading: boolean;
  selectedUser: User | null;
  showUserModal: boolean;
  showDiamondModal: boolean;
}

const UserManagement: React.FC = () => {
  const [state, setState] = useState<UserManagementState>({
    users: [],
    loading: true,
    selectedUser: null,
    showUserModal: false,
    showDiamondModal: false,
  });

  const [diamondAmount, setDiamondAmount] = useState<number>(0);
  const [diamondDescription, setDiamondDescription] = useState<string>("");
  const [diamondType, setDiamondType] = useState<"credit" | "debit">("credit");

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await userApi.getUsers();
      setState((prev) => ({
        ...prev,
        users: response.users || [],
        loading: false,
      }));
    } catch (error) {
      console.error("Failed to fetch users:", error);
      setState((prev) => ({ ...prev, loading: false }));
    }
  };

  const handleToggleUserStatus = async (user: User) => {
    try {
      await userApi.updateUser(user.id, {
        ...user,
        is_active: !user.is_active,
      });
      await fetchUsers();
    } catch (error) {
      console.error("Failed to update user:", error);
    }
  };

  const handleDeleteUser = async (userId: number) => {
    if (window.confirm("Are you sure you want to delete this user?")) {
      try {
        await userApi.deleteUser(userId);
        await fetchUsers();
      } catch (error) {
        console.error("Failed to delete user:", error);
      }
    }
  };

  const handleDiamondTransaction = async () => {
    if (!state.selectedUser || diamondAmount <= 0) return;

    try {
      if (diamondType === "credit") {
        await diamondApi.creditDiamonds({
          user_id: state.selectedUser.id,
          amount: diamondAmount,
          type: "credit",
          description:
            diamondDescription ||
            `Admin ${diamondType}: ${diamondAmount} diamonds`,
        });
      } else {
        await diamondApi.debitDiamonds({
          user_id: state.selectedUser.id,
          amount: diamondAmount,
          type: "debit",
          description:
            diamondDescription ||
            `Admin ${diamondType}: ${diamondAmount} diamonds`,
        });
      }

      setState((prev) => ({
        ...prev,
        showDiamondModal: false,
        selectedUser: null,
      }));
      setDiamondAmount(0);
      setDiamondDescription("");
      alert(
        `Successfully ${
          diamondType === "credit" ? "added" : "deducted"
        } ${diamondAmount} diamonds`
      );
    } catch (error) {
      console.error("Failed to process diamond transaction:", error);
      alert("Failed to process diamond transaction");
    }
  };

  if (state.loading) {
    return (
      <AdminLayout>
        <div className="flex justify-center items-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
            <p>Loading users...</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
          <div className="sm:flex sm:items-center">
            <div className="sm:flex-auto">
              <h1 className="text-2xl font-semibold text-gray-900">
                User Management
              </h1>
              <p className="mt-2 text-sm text-gray-700">
                Manage users, their status, and diamond balances.
              </p>
            </div>
            <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
              <button
                onClick={fetchUsers}
                className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto"
              >
                Refresh
              </button>
            </div>
          </div>

          {/* Users Table */}
          <div className="mt-8 flex flex-col">
            <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
              <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                  <table className="min-w-full divide-y divide-gray-300">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          User
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Status
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Roles
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Created
                        </th>
                        <th className="relative px-6 py-3">
                          <span className="sr-only">Actions</span>
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {state.users.map((user) => (
                        <tr key={user.id}>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <div className="flex items-center">
                              <div className="h-10 w-10 flex-shrink-0">
                                <div className="h-10 w-10 rounded-full bg-gray-300 flex items-center justify-center">
                                  <span className="text-sm font-medium text-gray-700">
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
                              </div>
                            </div>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                user.is_active
                                  ? "bg-green-100 text-green-800"
                                  : "bg-red-100 text-red-800"
                              }`}
                            >
                              {user.is_active ? "Active" : "Inactive"}
                            </span>
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {user.roles?.map((role) => role.name).join(", ") ||
                              "No roles"}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                            {new Date(user.created_at).toLocaleDateString()}
                          </td>
                          <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium space-x-2">
                            <button
                              onClick={() => handleToggleUserStatus(user)}
                              className={`inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded ${
                                user.is_active
                                  ? "bg-red-100 text-red-700 hover:bg-red-200"
                                  : "bg-green-100 text-green-700 hover:bg-green-200"
                              }`}
                            >
                              {user.is_active ? "Deactivate" : "Activate"}
                            </button>
                            <button
                              onClick={() =>
                                setState((prev) => ({
                                  ...prev,
                                  selectedUser: user,
                                  showDiamondModal: true,
                                }))
                              }
                              className="inline-flex items-center space-x-1 px-3 py-1 border border-transparent text-xs font-medium rounded bg-yellow-100 text-yellow-700 hover:bg-yellow-200"
                            >
                              <FaGem />
                              <span>Manage</span>
                            </button>
                            <button
                              onClick={() => handleDeleteUser(user.id)}
                              className="inline-flex items-center px-3 py-1 border border-transparent text-xs font-medium rounded bg-red-100 text-red-700 hover:bg-red-200"
                            >
                              Delete
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Diamond Management Modal */}
      {state.showDiamondModal && state.selectedUser && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 overflow-y-auto h-full w-full flex items-center justify-center z-50">
          <div className="bg-white p-6 rounded-lg shadow-xl w-96">
            <h3 className="text-lg font-medium text-gray-900 mb-4">
              Manage Diamonds for {state.selectedUser.username}
            </h3>

            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Transaction Type
                </label>
                <select
                  value={diamondType}
                  onChange={(e) =>
                    setDiamondType(e.target.value as "credit" | "debit")
                  }
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                >
                  <option value="credit">Add Diamonds (Credit)</option>
                  <option value="debit">Remove Diamonds (Debit)</option>
                </select>
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Amount
                </label>
                <input
                  type="number"
                  value={diamondAmount}
                  onChange={(e) => setDiamondAmount(Number(e.target.value))}
                  min="1"
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                  placeholder="Enter amount"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700">
                  Description
                </label>
                <input
                  type="text"
                  value={diamondDescription}
                  onChange={(e) => setDiamondDescription(e.target.value)}
                  className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                  placeholder="Transaction description (optional)"
                />
              </div>
            </div>

            <div className="mt-6 flex justify-end space-x-3">
              <button
                onClick={() =>
                  setState((prev) => ({
                    ...prev,
                    showDiamondModal: false,
                    selectedUser: null,
                  }))
                }
                className="px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 border border-gray-300 rounded-md hover:bg-gray-200"
              >
                Cancel
              </button>
              <button
                onClick={handleDiamondTransaction}
                disabled={diamondAmount <= 0}
                className="px-4 py-2 text-sm font-medium text-white bg-indigo-600 border border-transparent rounded-md hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {diamondType === "credit" ? "Add" : "Remove"} Diamonds
              </button>
            </div>
          </div>
        </div>
      )}
    </AdminLayout>
  );
};

export default UserManagement;
