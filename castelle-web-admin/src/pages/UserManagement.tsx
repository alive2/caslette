import React, { useEffect, useState } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import DataTable from "../components/common/DataTable";
import type {
  DataTableColumn,
  DataTableAction,
} from "../components/common/DataTable";
import { userApi, diamondApi, roleApi, permissionApi } from "../services/api";
import type { User, Role, Permission } from "../types";
import {
  FaGem,
  FaUserShield,
  FaTrash,
  FaCheck,
  FaTimes,
  FaUser,
  FaCrown,
  FaKey,
} from "react-icons/fa";
import { cosmic } from "../styles/cosmic-theme";

interface UserManagementState {
  users: User[];
  roles: Role[];
  permissions: Permission[];
  loading: boolean;
  selectedUser: User | null;
  showUserModal: boolean;
  showDiamondModal: boolean;
  showRoleModal: boolean;
  showPermissionModal: boolean;
}

const UserManagement: React.FC = () => {
  const [state, setState] = useState<UserManagementState>({
    users: [],
    roles: [],
    permissions: [],
    loading: true,
    selectedUser: null,
    showUserModal: false,
    showDiamondModal: false,
    showRoleModal: false,
    showPermissionModal: false,
  });

  const [diamondAmount, setDiamondAmount] = useState<number>(0);
  const [diamondDescription, setDiamondDescription] = useState<string>("");
  const [diamondType, setDiamondType] = useState<"credit" | "debit">("credit");
  const [selectedRoles, setSelectedRoles] = useState<number[]>([]);
  const [selectedPermissions, setSelectedPermissions] = useState<number[]>([]);

  useEffect(() => {
    fetchUsers();
    fetchRoles();
    fetchPermissions();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await userApi.getUsers();
      setState((prev) => ({
        ...prev,
        users: response.data?.users || [],
        loading: false,
      }));
    } catch (error) {
      console.error("Failed to fetch users:", error);
      setState((prev) => ({ ...prev, loading: false }));
    }
  };

  // Define DataTable columns
  const userColumns: DataTableColumn<User>[] = [
    {
      key: "username",
      title: "User",
      sortable: true,
      render: (_, user) => (
        <div className="flex items-center">
          <div className="flex-shrink-0 h-10 w-10">
            <div className="h-10 w-10 rounded-full bg-gradient-to-br from-purple-500 to-blue-500 flex items-center justify-center">
              <span className="text-white font-medium">
                {user.username.charAt(0).toUpperCase()}
              </span>
            </div>
          </div>
          <div className="ml-4">
            <div className="text-sm font-medium text-white">
              {user.username}
            </div>
            <div className="text-sm text-gray-400">
              {user.first_name} {user.last_name}
            </div>
          </div>
        </div>
      ),
      mobileLabel: "User",
    },
    {
      key: "email",
      title: "Email",
      sortable: true,
      className: "text-gray-300",
      mobileLabel: "Email",
    },
    {
      key: "is_active",
      title: "Status",
      sortable: true,
      render: (_, user) => (
        <span
          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
            user.is_active
              ? "bg-green-500/20 text-green-400 border border-green-500/30"
              : "bg-red-500/20 text-red-400 border border-red-500/30"
          }`}
        >
          {user.is_active ? "Active" : "Inactive"}
        </span>
      ),
      mobileLabel: "Status",
    },
    {
      key: "balance",
      title: "Balance",
      sortable: true,
      render: (_, user) => (
        <div className="flex items-center">
          <FaGem className="w-4 h-4 text-yellow-400 mr-2" />
          <span className="text-yellow-300 font-medium">
            {user.balance?.toLocaleString() || "0"}
          </span>
        </div>
      ),
      mobileLabel: "Balance",
    },
    {
      key: "roles",
      title: "Roles",
      render: (_, user) => (
        <div className="flex flex-wrap gap-1">
          {user.roles?.map((role) => (
            <span
              key={role.id}
              className="inline-flex items-center px-2 py-1 text-xs font-medium rounded-md bg-purple-500/20 text-purple-300 border border-purple-500/30"
            >
              <FaCrown className="mr-1 text-xs" />
              {role.name}
              <button
                onClick={() => handleRemoveRole(user.id, role.id)}
                className="ml-1 text-purple-400 hover:text-red-400 transition-colors"
              >
                <FaTimes className="text-xs" />
              </button>
            </span>
          )) || (
            <span className="text-gray-500 text-sm">No roles assigned</span>
          )}
        </div>
      ),
      mobileLabel: "Roles",
    },
  ];

  // Define DataTable actions
  const userActions: DataTableAction<User>[] = [
    {
      label: "Roles",
      icon: <FaUserShield />,
      onClick: (user) => openRoleModal(user),
      className: `${cosmic.button.secondary} text-xs`,
    },
    {
      label: "Permissions",
      icon: <FaKey />,
      onClick: (user) => openPermissionModal(user),
      className: `${cosmic.button.ghost} text-xs`,
    },
    {
      label: "Diamonds",
      icon: <FaGem />,
      onClick: (user) => openDiamondModal(user),
      className: `${cosmic.button.primary} text-xs`,
    },
    {
      label: "Activate/Deactivate",
      icon: <FaCheck />,
      onClick: (user) => handleToggleUserStatus(user),
      className: `${cosmic.button.secondary} text-xs`,
    },
    {
      label: "Delete",
      icon: <FaTrash />,
      onClick: (user) => handleDeleteUser(user.id),
      className: `${cosmic.button.danger} text-xs`,
    },
  ];

  const fetchRoles = async () => {
    try {
      const response = await roleApi.getRoles();
      setState((prev) => ({
        ...prev,
        roles: response.data?.roles || [],
      }));
    } catch (error) {
      console.error("Failed to fetch roles:", error);
    }
  };

  const fetchPermissions = async () => {
    try {
      const response = await permissionApi.getPermissions();
      setState((prev) => ({
        ...prev,
        permissions: response.data?.permissions || [],
      }));
    } catch (error) {
      console.error("Failed to fetch permissions:", error);
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

  const handleAssignRoles = async () => {
    if (!state.selectedUser || selectedRoles.length === 0) return;

    try {
      await userApi.assignRoles(state.selectedUser.id, selectedRoles);
      setState((prev) => ({
        ...prev,
        showRoleModal: false,
        selectedUser: null,
      }));
      setSelectedRoles([]);
      await fetchUsers();
      alert("Roles assigned successfully!");
    } catch (error) {
      console.error("Failed to assign roles:", error);
      alert("Failed to assign roles");
    }
  };

  const handleRemoveRole = async (userId: number, roleId: number) => {
    try {
      await userApi.removeRole(userId, roleId);
      await fetchUsers();
    } catch (error) {
      console.error("Failed to remove role:", error);
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

  const openRoleModal = (user: User) => {
    setState((prev) => ({
      ...prev,
      selectedUser: user,
      showRoleModal: true,
    }));
    setSelectedRoles(user.roles?.map((role) => role.id) || []);
  };

  const openDiamondModal = (user: User) => {
    setState((prev) => ({
      ...prev,
      selectedUser: user,
      showDiamondModal: true,
    }));
  };

  const openPermissionModal = (user: User) => {
    setState((prev) => ({
      ...prev,
      selectedUser: user,
      showPermissionModal: true,
    }));
    setSelectedPermissions(
      user.permissions?.map((permission) => permission.id) || []
    );
  };

  const handleAssignPermissions = async () => {
    if (!state.selectedUser) return;

    try {
      // First, get current user permissions to remove them
      const currentPermissions = state.selectedUser.permissions || [];

      // Remove all existing permissions
      for (const permission of currentPermissions) {
        await userApi.removePermission(state.selectedUser.id, permission.id);
      }

      // Then assign the new permissions (only if there are any selected)
      if (selectedPermissions.length > 0) {
        await userApi.assignPermissions(
          state.selectedUser.id,
          selectedPermissions
        );
      }

      setState((prev) => ({
        ...prev,
        showPermissionModal: false,
      }));

      // Refresh users to get updated data
      fetchUsers();
      alert("User permissions updated successfully!");
    } catch (error) {
      console.error("Failed to assign permissions:", error);
      alert("Failed to assign permissions");
    }
  };

  if (state.loading) {
    return (
      <AdminLayout>
        <div className="flex justify-center items-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-500 mx-auto mb-4"></div>
            <p className="text-white">Loading users...</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  return (
    <AdminLayout>
      <div className="p-6">
        <div className="max-w-7xl mx-auto">
          <div className="sm:flex sm:items-center mb-8">
            <div className="sm:flex-auto">
              <h1 className="text-3xl font-bold text-white flex items-center space-x-3">
                <FaUser className="text-purple-400" />
                <span>User Management</span>
              </h1>
              <p className="mt-2 text-gray-300">
                Manage users, their roles, and diamond balances
              </p>
            </div>
          </div>

          <DataTable
            data={state.users}
            columns={userColumns}
            actions={userActions}
            loading={state.loading}
            searchable={true}
            emptyMessage="No users found"
            cardTitleExtractor={(user) => user.username}
            cardSubtitleExtractor={(user) =>
              `${user.first_name} ${user.last_name}`
            }
            cardKeyExtractor={(user) => user.id.toString()}
            useDropdownActions={true}
          />
        </div>

        {/* Role Assignment Modal */}
        {state.showRoleModal && state.selectedUser && (
          <div className="fixed inset-0 bg-black/75 backdrop-blur-sm flex items-center justify-center z-50">
            <div
              className={`${cosmic.cardElevated} max-w-xl w-full mx-4 max-h-[90vh] flex flex-col`}
            >
              {/* Modal Header */}
              <div className="flex items-center justify-between p-6 border-b border-white/10">
                <div>
                  <h3 className="text-xl font-semibold text-white flex items-center">
                    <FaUserShield className="mr-3 text-blue-400" />
                    Manage User Roles
                  </h3>
                  <p className="text-gray-300 mt-1">
                    Assign roles to{" "}
                    <span className="font-medium text-blue-300">
                      {state.selectedUser.username}
                    </span>
                  </p>
                </div>
                <button
                  onClick={() =>
                    setState((prev) => ({ ...prev, showRoleModal: false }))
                  }
                  className="text-gray-400 hover:text-white transition-colors p-1"
                >
                  <FaTimes className="text-xl" />
                </button>
              </div>

              {/* Modal Body */}
              <div className="flex-1 overflow-hidden p-6">
                {/* Current User Roles Summary */}
                <div className="mb-6 p-4 bg-blue-500/10 border border-blue-500/20 rounded-lg">
                  <h4 className="text-sm font-medium text-blue-300 mb-2 flex items-center">
                    <FaCrown className="mr-2" />
                    Current Roles ({selectedRoles.length} selected)
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {selectedRoles.length > 0 ? (
                      state.roles
                        .filter((r) => selectedRoles.includes(r.id))
                        .map((role) => (
                          <span
                            key={role.id}
                            className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-blue-500/20 text-blue-300 border border-blue-500/30"
                          >
                            <FaCrown className="mr-1 text-xs" />
                            {role.name}
                            <button
                              onClick={() =>
                                setSelectedRoles((prev) =>
                                  prev.filter((id) => id !== role.id)
                                )
                              }
                              className="ml-2 text-blue-400 hover:text-blue-200 transition-colors"
                            >
                              <FaTimes className="text-xs" />
                            </button>
                          </span>
                        ))
                    ) : (
                      <span className="text-gray-400 text-sm">
                        No roles selected
                      </span>
                    )}
                  </div>
                </div>

                {/* Available Roles */}
                <div className="space-y-4">
                  <h4 className="text-sm font-medium text-gray-300 flex items-center">
                    <FaCheck className="mr-2" />
                    Available System Roles
                  </h4>

                  {/* Role Cards */}
                  <div className="space-y-3 max-h-64 overflow-y-auto pr-2">
                    {state.roles.map((role) => (
                      <label
                        key={role.id}
                        className={`flex items-start space-x-3 p-4 rounded-lg cursor-pointer transition-all ${
                          selectedRoles.includes(role.id)
                            ? "bg-blue-500/20 border-2 border-blue-500/40"
                            : "bg-white/5 border-2 border-white/10 hover:bg-white/10"
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={selectedRoles.includes(role.id)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedRoles([...selectedRoles, role.id]);
                            } else {
                              setSelectedRoles(
                                selectedRoles.filter((id) => id !== role.id)
                              );
                            }
                          }}
                          className="mt-1 rounded bg-white/10 border-white/20 text-blue-600 focus:ring-blue-500 focus:ring-offset-0"
                        />
                        <div className="flex-1 min-w-0">
                          <div className="flex items-center space-x-2 mb-2">
                            <FaCrown
                              className={`text-sm ${
                                selectedRoles.includes(role.id)
                                  ? "text-blue-400"
                                  : "text-gray-400"
                              }`}
                            />
                            <span className="text-white font-medium text-lg">
                              {role.name}
                            </span>
                            <span
                              className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                                role.name.toLowerCase() === "admin"
                                  ? "bg-red-700 text-red-200"
                                  : role.name.toLowerCase() === "manager"
                                  ? "bg-yellow-700 text-yellow-200"
                                  : "bg-gray-700 text-gray-300"
                              }`}
                            >
                              {role.name.toLowerCase()}
                            </span>
                          </div>
                          <p className="text-gray-400 text-sm leading-relaxed">
                            {role.description}
                          </p>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>
              </div>

              {/* Modal Footer */}
              <div className="flex items-center justify-between p-6 border-t border-white/10 bg-white/5">
                <div className="text-sm text-gray-400">
                  {selectedRoles.length} of {state.roles.length} roles selected
                </div>
                <div className="flex space-x-3">
                  <button
                    onClick={() =>
                      setState((prev) => ({ ...prev, showRoleModal: false }))
                    }
                    className={cosmic.button.secondary}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleAssignRoles}
                    className={`${cosmic.button.primary} flex items-center`}
                  >
                    <FaCheck className="mr-2" />
                    Apply Roles ({selectedRoles.length})
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Diamond Transaction Modal */}
        {state.showDiamondModal && state.selectedUser && (
          <div className="fixed inset-0 bg-black/75 backdrop-blur-sm flex items-center justify-center z-50">
            <div
              className={`${cosmic.cardElevated} max-w-2xl w-full mx-4 max-h-[90vh] flex flex-col`}
            >
              {/* Modal Header */}
              <div className="flex items-center justify-between p-6 border-b border-white/10">
                <div>
                  <h3 className="text-xl font-semibold text-white flex items-center">
                    <FaGem className="mr-3 text-yellow-400" />
                    Diamond Transaction
                  </h3>
                  <p className="text-gray-300 mt-1">
                    Manage diamonds for{" "}
                    <span className="font-medium text-yellow-300">
                      {state.selectedUser.username}
                    </span>
                  </p>
                </div>
                <button
                  onClick={() =>
                    setState((prev) => ({ ...prev, showDiamondModal: false }))
                  }
                  className="text-gray-400 hover:text-white transition-colors p-1"
                >
                  <FaTimes className="text-xl" />
                </button>
              </div>

              {/* Modal Body */}
              <div className="flex-1 overflow-hidden p-6">
                <div className="space-y-6">
                  {/* Transaction Type Selection */}
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-3">
                      Transaction Type
                    </label>
                    <div className="grid grid-cols-2 gap-3">
                      <button
                        onClick={() => setDiamondType("credit")}
                        className={`p-4 rounded-lg border-2 transition-all ${
                          diamondType === "credit"
                            ? "border-green-500 bg-green-500/20 text-green-300"
                            : "border-white/20 bg-white/5 text-gray-300 hover:bg-white/10"
                        }`}
                      >
                        <div className="flex items-center space-x-3">
                          <div
                            className={`w-3 h-3 rounded-full ${
                              diamondType === "credit"
                                ? "bg-green-400"
                                : "bg-gray-400"
                            }`}
                          ></div>
                          <div className="text-left">
                            <div className="font-medium">Credit</div>
                            <div className="text-xs opacity-75">
                              Add diamonds
                            </div>
                          </div>
                        </div>
                      </button>
                      <button
                        onClick={() => setDiamondType("debit")}
                        className={`p-4 rounded-lg border-2 transition-all ${
                          diamondType === "debit"
                            ? "border-red-500 bg-red-500/20 text-red-300"
                            : "border-white/20 bg-white/5 text-gray-300 hover:bg-white/10"
                        }`}
                      >
                        <div className="flex items-center space-x-3">
                          <div
                            className={`w-3 h-3 rounded-full ${
                              diamondType === "debit"
                                ? "bg-red-400"
                                : "bg-gray-400"
                            }`}
                          ></div>
                          <div className="text-left">
                            <div className="font-medium">Debit</div>
                            <div className="text-xs opacity-75">
                              Remove diamonds
                            </div>
                          </div>
                        </div>
                      </button>
                    </div>
                  </div>

                  {/* Amount Input */}
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-3">
                      Diamond Amount
                    </label>
                    <div className="relative">
                      <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                        <FaGem className="h-4 w-4 text-yellow-400" />
                      </div>
                      <input
                        type="number"
                        value={diamondAmount}
                        onChange={(e) =>
                          setDiamondAmount(Number(e.target.value))
                        }
                        className={`${cosmic.input} pl-10`}
                        placeholder="Enter amount"
                        min="1"
                      />
                    </div>
                    <p className="text-xs text-gray-400 mt-2">
                      {diamondType === "credit"
                        ? "Diamonds will be added to the user's account"
                        : "Diamonds will be deducted from the user's account"}
                    </p>
                  </div>

                  {/* Description */}
                  <div>
                    <label className="block text-sm font-medium text-gray-300 mb-3">
                      Transaction Description
                    </label>
                    <textarea
                      value={diamondDescription}
                      onChange={(e) => setDiamondDescription(e.target.value)}
                      className={`${cosmic.input} w-full resize-none`}
                      placeholder="Optional: Reason for this transaction..."
                      rows={4}
                    />
                    <p className="text-xs text-gray-400 mt-2">
                      This will appear in the transaction history
                    </p>
                  </div>

                  {/* Transaction Summary */}
                  {diamondAmount > 0 && (
                    <div
                      className={`p-4 rounded-lg border ${
                        diamondType === "credit"
                          ? "border-green-500/30 bg-green-500/10"
                          : "border-red-500/30 bg-red-500/10"
                      }`}
                    >
                      <h4 className="text-sm font-medium text-white mb-2 flex items-center">
                        <FaCheck className="mr-2 text-green-400" />
                        Transaction Summary
                      </h4>
                      <div className="space-y-1 text-sm">
                        <div className="flex justify-between">
                          <span className="text-gray-300">User:</span>
                          <span className="text-white font-medium">
                            {state.selectedUser.username}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-300">Type:</span>
                          <span
                            className={`font-medium ${
                              diamondType === "credit"
                                ? "text-green-300"
                                : "text-red-300"
                            }`}
                          >
                            {diamondType === "credit"
                              ? "Credit (+)"
                              : "Debit (-)"}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-300">Amount:</span>
                          <span className="text-yellow-300 font-medium flex items-center">
                            <FaGem className="mr-1 text-xs" />
                            {diamondAmount.toLocaleString()}
                          </span>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              </div>

              {/* Modal Footer */}
              <div className="flex items-center justify-between p-6 border-t border-white/10 bg-white/5">
                <div className="text-sm text-gray-400">
                  {diamondAmount > 0
                    ? `Processing ${diamondAmount.toLocaleString()} diamonds`
                    : "Enter amount to proceed"}
                </div>
                <div className="flex space-x-3">
                  <button
                    onClick={() =>
                      setState((prev) => ({ ...prev, showDiamondModal: false }))
                    }
                    className={cosmic.button.secondary}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleDiamondTransaction}
                    className={`${cosmic.button.primary} flex items-center px-6 py-3 text-lg font-semibold bg-gradient-to-r from-yellow-600 to-yellow-500 hover:from-yellow-500 hover:to-yellow-400 transition-all duration-200 shadow-lg hover:shadow-xl disabled:opacity-50 disabled:cursor-not-allowed`}
                    disabled={diamondAmount <= 0}
                  >
                    <FaGem className="mr-3 text-xl" />
                    Process Transaction
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* User Permission Assignment Modal */}
        {state.showPermissionModal && state.selectedUser && (
          <div className="fixed inset-0 bg-black/75 backdrop-blur-sm flex items-center justify-center z-50">
            <div
              className={`${cosmic.cardElevated} max-w-2xl w-full mx-4 max-h-[90vh] flex flex-col`}
            >
              {/* Modal Header */}
              <div className="flex items-center justify-between p-6 border-b border-white/10">
                <div>
                  <h3 className="text-xl font-semibold text-white flex items-center">
                    <FaKey className="mr-3 text-purple-400" />
                    Manage Permissions
                  </h3>
                  <p className="text-gray-300 mt-1">
                    Assign permissions to{" "}
                    <span className="font-medium text-purple-300">
                      {state.selectedUser.username}
                    </span>
                  </p>
                </div>
                <button
                  onClick={() =>
                    setState((prev) => ({
                      ...prev,
                      showPermissionModal: false,
                    }))
                  }
                  className="text-gray-400 hover:text-white transition-colors p-1"
                >
                  <FaTimes className="text-xl" />
                </button>
              </div>

              {/* Modal Body */}
              <div className="flex-1 overflow-hidden p-6">
                {/* Current User Permissions Summary */}
                <div className="mb-6 p-4 bg-purple-500/10 border border-purple-500/20 rounded-lg">
                  <h4 className="text-sm font-medium text-purple-300 mb-2 flex items-center">
                    <FaUserShield className="mr-2" />
                    Current Permissions ({selectedPermissions.length} selected)
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {selectedPermissions.length > 0 ? (
                      state.permissions
                        .filter((p) => selectedPermissions.includes(p.id))
                        .map((permission) => (
                          <span
                            key={permission.id}
                            className="inline-flex items-center px-3 py-1 rounded-full text-xs font-medium bg-purple-500/20 text-purple-300 border border-purple-500/30"
                          >
                            {permission.name}
                            <button
                              onClick={() =>
                                setSelectedPermissions((prev) =>
                                  prev.filter((id) => id !== permission.id)
                                )
                              }
                              className="ml-2 text-purple-400 hover:text-purple-200 transition-colors"
                            >
                              <FaTimes className="text-xs" />
                            </button>
                          </span>
                        ))
                    ) : (
                      <span className="text-gray-400 text-sm">
                        No permissions selected
                      </span>
                    )}
                  </div>
                </div>

                {/* Available Permissions */}
                <div className="space-y-4">
                  <h4 className="text-sm font-medium text-gray-300 flex items-center">
                    <FaCheck className="mr-2" />
                    Available Permissions
                  </h4>

                  {/* Permission Groups */}
                  <div className="space-y-3 max-h-96 overflow-y-auto pr-2">
                    {/* Group permissions by resource */}
                    {Object.entries(
                      state.permissions.reduce((groups, permission) => {
                        const resource = permission.resource || "general";
                        if (!groups[resource]) groups[resource] = [];
                        groups[resource].push(permission);
                        return groups;
                      }, {} as Record<string, typeof state.permissions>)
                    ).map(([resource, permissions]) => (
                      <div
                        key={resource}
                        className="bg-white/5 rounded-lg p-4 border border-white/10"
                      >
                        <h5 className="text-white font-medium mb-3 capitalize flex items-center">
                          <div className="w-2 h-2 bg-purple-400 rounded-full mr-2"></div>
                          {resource} Permissions
                        </h5>
                        <div className="grid grid-cols-1 gap-2">
                          {permissions.map((permission) => (
                            <label
                              key={permission.id}
                              className={`flex items-start space-x-3 p-3 rounded-lg cursor-pointer transition-all ${
                                selectedPermissions.includes(permission.id)
                                  ? "bg-purple-500/20 border border-purple-500/40"
                                  : "bg-white/5 border border-white/10 hover:bg-white/10"
                              }`}
                            >
                              <input
                                type="checkbox"
                                checked={selectedPermissions.includes(
                                  permission.id
                                )}
                                onChange={(e) => {
                                  if (e.target.checked) {
                                    setSelectedPermissions([
                                      ...selectedPermissions,
                                      permission.id,
                                    ]);
                                  } else {
                                    setSelectedPermissions(
                                      selectedPermissions.filter(
                                        (id) => id !== permission.id
                                      )
                                    );
                                  }
                                }}
                                className="mt-1 rounded bg-white/10 border-white/20 text-purple-600 focus:ring-purple-500 focus:ring-offset-0"
                              />
                              <div className="flex-1 min-w-0">
                                <div className="flex items-center space-x-2">
                                  <span className="text-white font-medium">
                                    {permission.name}
                                  </span>
                                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-gray-700 text-gray-300">
                                    {permission.action}
                                  </span>
                                </div>
                                <p className="text-gray-400 text-sm mt-1 leading-relaxed">
                                  {permission.description}
                                </p>
                              </div>
                            </label>
                          ))}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              </div>

              {/* Modal Footer */}
              <div className="flex items-center justify-between p-6 border-t border-white/10 bg-white/5">
                <div className="text-sm text-gray-400">
                  {selectedPermissions.length} of {state.permissions.length}{" "}
                  permissions selected
                </div>
                <div className="flex space-x-3">
                  <button
                    onClick={() =>
                      setState((prev) => ({
                        ...prev,
                        showPermissionModal: false,
                      }))
                    }
                    className={cosmic.button.secondary}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleAssignPermissions}
                    className={`${cosmic.button.primary} flex items-center`}
                  >
                    <FaCheck className="mr-2" />
                    Apply Permissions ({selectedPermissions.length})
                  </button>
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </AdminLayout>
  );
};

export default UserManagement;
