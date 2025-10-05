import React, { useEffect, useState } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import { roleApi, permissionApi } from "../services/api";
import type { Role, Permission, CreateRoleRequest } from "../types";
import {
  FaPlus,
  FaEdit,
  FaTrash,
  FaCheck,
  FaTimes,
  FaKey,
  FaCrown,
  FaUsers,
  FaShieldAlt,
} from "react-icons/fa";
import { cosmic } from "../styles/cosmic-theme";

interface RoleManagementState {
  roles: Role[];
  permissions: Permission[];
  loading: boolean;
  selectedRole: Role | null;
  showRoleModal: boolean;
  showRolePermissionModal: boolean;
}

const RoleManagement: React.FC = () => {
  const [state, setState] = useState<RoleManagementState>({
    roles: [],
    permissions: [],
    loading: true,
    selectedRole: null,
    showRoleModal: false,
    showRolePermissionModal: false,
  });

  const [roleForm, setRoleForm] = useState<CreateRoleRequest>({
    name: "",
    description: "",
  });

  const [selectedPermissions, setSelectedPermissions] = useState<number[]>([]);
  const [editingRole, setEditingRole] = useState<Role | null>(null);

  useEffect(() => {
    fetchRoles();
    fetchPermissions();
  }, []);

  const fetchRoles = async () => {
    try {
      console.log("Fetching roles...");
      const response = await roleApi.getRoles();
      console.log("Roles response:", response);
      setState((prev) => ({
        ...prev,
        roles: response.data?.roles || [],
        loading: false,
      }));
    } catch (error) {
      console.error("Failed to fetch roles:", error);
      setState((prev) => ({ ...prev, loading: false }));
    }
  };

  const fetchPermissions = async () => {
    try {
      console.log("Fetching permissions...");
      const response = await permissionApi.getPermissions();
      console.log("Permissions response:", response);
      setState((prev) => ({
        ...prev,
        permissions: response.data?.permissions || [],
      }));
    } catch (error) {
      console.error("Failed to fetch permissions:", error);
    }
  };

  const handleCreateRole = async () => {
    if (!roleForm.name.trim()) return;

    try {
      if (editingRole) {
        await roleApi.updateRole(editingRole.id, roleForm);
      } else {
        await roleApi.createRole(roleForm);
      }

      setState((prev) => ({ ...prev, showRoleModal: false }));
      setRoleForm({ name: "", description: "" });
      setEditingRole(null);
      await fetchRoles();
      alert(`Role ${editingRole ? "updated" : "created"} successfully!`);
    } catch (error) {
      console.error("Failed to save role:", error);
      alert("Failed to save role");
    }
  };

  const handleDeleteRole = async (roleId: number) => {
    if (window.confirm("Are you sure you want to delete this role?")) {
      try {
        await roleApi.deleteRole(roleId);
        await fetchRoles();
        alert("Role deleted successfully!");
      } catch (error) {
        console.error("Failed to delete role:", error);
        alert("Failed to delete role");
      }
    }
  };

  const handleAssignPermissions = async () => {
    if (!state.selectedRole || selectedPermissions.length === 0) return;

    try {
      await roleApi.assignPermissions(
        state.selectedRole.id,
        selectedPermissions
      );
      setState((prev) => ({
        ...prev,
        showRolePermissionModal: false,
        selectedRole: null,
      }));
      setSelectedPermissions([]);
      await fetchRoles();
      alert("Permissions assigned successfully!");
    } catch (error) {
      console.error("Failed to assign permissions:", error);
      alert("Failed to assign permissions");
    }
  };

  const openRoleModal = (role?: Role) => {
    if (role) {
      setEditingRole(role);
      setRoleForm({ name: role.name, description: role.description });
    } else {
      setEditingRole(null);
      setRoleForm({ name: "", description: "" });
    }
    setState((prev) => ({ ...prev, showRoleModal: true }));
  };

  const openRolePermissionModal = (role: Role) => {
    setState((prev) => ({
      ...prev,
      selectedRole: role,
      showRolePermissionModal: true,
    }));
    setSelectedPermissions(role.permissions?.map((p) => p.id) || []);
  };

  return (
    <AdminLayout>
      <div className="max-w-7xl mx-auto p-6 space-y-6">
        {/* Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
          <div>
            <h1 className="text-3xl font-bold text-white flex items-center gap-3">
              <FaShieldAlt className="text-purple-400" />
              <span>Role Management</span>
            </h1>
            <p className="mt-2 text-gray-300">
              Manage roles and assign permissions from the server
            </p>
          </div>
          <button
            onClick={() => openRoleModal()}
            className={`${cosmic.button.primary} flex items-center gap-2 px-6 py-3`}
          >
            <FaPlus />
            New Role
          </button>
        </div>

        {/* Roles Section */}
        <div>
          <div className="mb-6">
            <h2 className="text-2xl font-semibold text-white flex items-center gap-3">
              <FaCrown className="text-yellow-400" />
              <span>Roles</span>
              <span className="text-lg text-gray-400 font-normal">
                ({state.roles.length})
              </span>
            </h2>
          </div>

          {state.loading ? (
            <div className="flex flex-col items-center justify-center py-16">
              <div className="inline-block animate-spin rounded-full h-12 w-12 border-4 border-purple-400 border-t-transparent"></div>
              <p className="mt-4 text-gray-400 text-lg">Loading roles...</p>
            </div>
          ) : state.roles.length === 0 ? (
            <div className={`${cosmic.cardElevated} text-center py-16`}>
              <FaUsers className="mx-auto text-gray-400 text-6xl mb-6" />
              <h3 className="text-xl font-semibold text-white mb-2">
                No roles found
              </h3>
              <p className="text-gray-400">
                Create your first role to get started
              </p>
            </div>
          ) : (
            <div className="space-y-6">
              {state.roles.map((role) => (
                <div
                  key={role.id}
                  className={`${cosmic.cardElevated} hover:bg-white/[0.08] transition-all duration-200 group`}
                >
                  <div className="p-6">
                    <div className="flex items-start justify-between">
                      <div className="flex-1">
                        <div className="flex items-center gap-3 mb-3">
                          <div className="p-2 bg-purple-500/20 rounded-lg group-hover:bg-purple-500/30 transition-colors">
                            <FaShieldAlt className="text-purple-400 text-lg" />
                          </div>
                          <div className="flex items-center gap-3">
                            <h3 className="text-xl font-semibold text-white">
                              {role.name}
                            </h3>
                            <span className="inline-flex items-center px-3 py-1 text-sm font-medium rounded-full bg-purple-500/20 text-purple-300 border border-purple-500/30">
                              {role.permissions?.length || 0} permissions
                            </span>
                          </div>
                        </div>

                        <p className="text-gray-400 mb-4 leading-relaxed">
                          {role.description}
                        </p>

                        {/* Permissions Preview */}
                        {role.permissions && role.permissions.length > 0 && (
                          <div className="flex flex-wrap gap-2">
                            {role.permissions.slice(0, 4).map((permission) => (
                              <span
                                key={permission.id}
                                className="inline-flex items-center px-3 py-1 text-sm font-medium rounded-lg bg-green-500/20 text-green-300 border border-green-500/30"
                              >
                                <FaKey className="w-3 h-3 mr-2" />
                                {permission.name}
                              </span>
                            ))}
                            {role.permissions.length > 4 && (
                              <span className="inline-flex items-center px-3 py-1 text-sm font-medium rounded-lg bg-gray-500/20 text-gray-300 border border-gray-500/30">
                                +{role.permissions.length - 4} more
                              </span>
                            )}
                          </div>
                        )}
                      </div>

                      {/* Actions */}
                      <div className="flex items-center gap-2 ml-6">
                        <button
                          onClick={() => openRolePermissionModal(role)}
                          className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg bg-purple-500/20 text-purple-300 border border-purple-500/30 hover:bg-purple-500/30 transition-colors"
                          title="Manage Permissions"
                        >
                          <FaKey className="w-4 h-4" />
                          <span className="hidden sm:inline">Permissions</span>
                        </button>
                        <button
                          onClick={() => openRoleModal(role)}
                          className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg bg-blue-500/20 text-blue-300 border border-blue-500/30 hover:bg-blue-500/30 transition-colors"
                          title="Edit Role"
                        >
                          <FaEdit className="w-4 h-4" />
                          <span className="hidden sm:inline">Edit</span>
                        </button>
                        <button
                          onClick={() => handleDeleteRole(role.id)}
                          className="flex items-center gap-2 px-4 py-2 text-sm font-medium rounded-lg bg-red-500/20 text-red-300 border border-red-500/30 hover:bg-red-500/30 transition-colors"
                          title="Delete Role"
                        >
                          <FaTrash className="w-4 h-4" />
                          <span className="hidden sm:inline">Delete</span>
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}

          {/* Role Modal - Professional Redesign */}
          {state.showRoleModal && (
            <div className="fixed inset-0 bg-black/80 backdrop-blur-sm flex items-center justify-center z-50 p-4">
              <div
                className={`${cosmic.cardElevated} max-w-2xl w-full relative`}
              >
                {/* Header */}
                <div className="flex items-center justify-between p-8 border-b border-white/10">
                  <div className="flex items-center space-x-4">
                    <div className="flex items-center justify-center w-12 h-12 bg-purple-500/20 rounded-xl">
                      <FaShieldAlt className="text-purple-400 text-xl" />
                    </div>
                    <div>
                      <h3 className="text-2xl font-bold text-white">
                        {editingRole ? "Edit Role" : "Create New Role"}
                      </h3>
                      <p className="text-gray-400 mt-1">
                        {editingRole
                          ? "Update role information and permissions"
                          : "Define a new role with specific permissions"}
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() =>
                      setState((prev) => ({ ...prev, showRoleModal: false }))
                    }
                    className="text-gray-400 hover:text-white p-3 hover:bg-white/10 rounded-xl transition-colors"
                  >
                    <FaTimes className="w-6 h-6" />
                  </button>
                </div>

                {/* Form */}
                <div className="p-8 space-y-8">
                  <div>
                    <label className="block text-lg font-semibold text-gray-200 mb-4">
                      Role Name <span className="text-red-400">*</span>
                    </label>
                    <input
                      type="text"
                      value={roleForm.name}
                      onChange={(e) =>
                        setRoleForm({ ...roleForm, name: e.target.value })
                      }
                      className={`${cosmic.input} text-xl py-4 px-6 w-full`}
                      placeholder="e.g., Content Manager, Administrator, Moderator"
                    />
                  </div>

                  <div>
                    <label className="block text-lg font-semibold text-gray-200 mb-4">
                      Description
                    </label>
                    <textarea
                      value={roleForm.description}
                      onChange={(e) =>
                        setRoleForm({
                          ...roleForm,
                          description: e.target.value,
                        })
                      }
                      className={`${cosmic.input} resize-none w-full py-4 px-6 text-lg leading-relaxed cosmic-scroll`}
                      placeholder="Describe the responsibilities, scope, and purpose of this role. What permissions should this role have? What can users with this role do in the system?"
                      rows={6}
                    />
                  </div>

                  {/* Enhanced Preview */}
                  {roleForm.name && (
                    <div className="bg-gradient-to-r from-purple-500/10 to-blue-500/10 border border-purple-500/20 rounded-xl p-6">
                      <h4 className="text-lg font-semibold text-purple-300 mb-4">
                        Role Preview
                      </h4>
                      <div className="flex items-center space-x-3 mb-3">
                        <div className="p-2 bg-purple-500/20 rounded-lg">
                          <FaShieldAlt className="text-purple-400 text-lg" />
                        </div>
                        <div>
                          <h5 className="text-xl font-bold text-white">
                            {roleForm.name}
                          </h5>
                          <span className="inline-flex items-center px-3 py-1 text-sm font-medium rounded-full bg-purple-500/20 text-purple-300 border border-purple-500/30">
                            Role
                          </span>
                        </div>
                      </div>
                      {roleForm.description && (
                        <p className="text-gray-300 leading-relaxed bg-black/20 p-4 rounded-lg">
                          {roleForm.description}
                        </p>
                      )}
                    </div>
                  )}
                </div>

                {/* Actions */}
                <div className="flex items-center justify-end space-x-4 p-8 border-t border-white/10 bg-white/5">
                  <button
                    onClick={() =>
                      setState((prev) => ({ ...prev, showRoleModal: false }))
                    }
                    className={`${cosmic.button.secondary} px-8 py-3 text-lg`}
                  >
                    Cancel
                  </button>
                  <button
                    onClick={handleCreateRole}
                    disabled={!roleForm.name.trim()}
                    className={`${cosmic.button.primary} px-8 py-3 text-lg flex items-center space-x-3 disabled:opacity-50 disabled:cursor-not-allowed`}
                  >
                    <FaCheck className="w-5 h-5" />
                    <span>{editingRole ? "Update Role" : "Create Role"}</span>
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Permission Assignment Modal - Redesigned */}
          {state.showRolePermissionModal && state.selectedRole && (
            <div className="fixed inset-0 bg-black/75 backdrop-blur-sm flex items-center justify-center z-50 p-4">
              <div
                className={`${cosmic.cardElevated} max-w-2xl w-full relative max-h-[90vh] flex flex-col`}
              >
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b border-white/10 flex-shrink-0">
                  <div className="flex items-center space-x-3">
                    <div className="flex items-center justify-center w-10 h-10 bg-green-500/20 rounded-lg">
                      <FaKey className="text-green-400" />
                    </div>
                    <div>
                      <h3 className="text-xl font-semibold text-white">
                        Manage Permissions
                      </h3>
                      <p className="text-sm text-gray-400">
                        Assign permissions to{" "}
                        <span className="font-medium text-purple-300">
                          {state.selectedRole.name}
                        </span>
                      </p>
                    </div>
                  </div>
                  <button
                    onClick={() =>
                      setState((prev) => ({
                        ...prev,
                        showRolePermissionModal: false,
                      }))
                    }
                    className="text-gray-400 hover:text-white p-2 hover:bg-white/10 rounded-lg transition-colors"
                  >
                    <FaTimes className="w-5 h-5" />
                  </button>
                </div>

                {/* Current Permissions Summary */}
                <div className="p-6 border-b border-white/10 bg-white/5 flex-shrink-0">
                  <h4 className="text-sm font-semibold text-gray-200 mb-3">
                    Current Permissions ({selectedPermissions.length})
                  </h4>
                  <div className="flex flex-wrap gap-2">
                    {selectedPermissions.length === 0 ? (
                      <span className="text-sm text-gray-500 italic">
                        No permissions assigned
                      </span>
                    ) : (
                      selectedPermissions.map((permId) => {
                        const permission = state.permissions.find(
                          (p) => p.id === permId
                        );
                        return permission ? (
                          <span
                            key={permission.id}
                            className="inline-flex items-center px-2 py-1 text-xs font-medium rounded-md bg-green-500/20 text-green-300 border border-green-500/30"
                          >
                            {permission.name}
                          </span>
                        ) : null;
                      })
                    )}
                  </div>
                </div>

                {/* Permissions Grid */}
                <div className="flex-1 overflow-y-auto p-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    {state.permissions.map((permission) => {
                      const isSelected = selectedPermissions.includes(
                        permission.id
                      );
                      return (
                        <label
                          key={permission.id}
                          className={`
                          flex items-start space-x-3 p-4 rounded-lg border-2 cursor-pointer transition-all
                          ${
                            isSelected
                              ? "border-green-500/50 bg-green-500/10"
                              : "border-white/10 bg-white/5 hover:border-white/20 hover:bg-white/10"
                          }
                        `}
                        >
                          <input
                            type="checkbox"
                            checked={isSelected}
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
                            className="mt-1 rounded bg-white/10 border-white/20 text-green-600 focus:ring-green-500 focus:ring-2"
                          />
                          <div className="flex-1">
                            <div className="flex items-center space-x-2 mb-2">
                              <h5 className="text-white font-medium">
                                {permission.name}
                              </h5>
                              {isSelected && (
                                <FaCheck className="text-green-400 text-sm" />
                              )}
                            </div>
                            <p className="text-gray-400 text-sm mb-2">
                              {permission.description}
                            </p>
                            <div className="flex space-x-2">
                              <span className="inline-flex items-center px-2 py-0.5 text-xs font-medium rounded bg-purple-500/20 text-purple-300">
                                {permission.resource}
                              </span>
                              <span className="inline-flex items-center px-2 py-0.5 text-xs font-medium rounded bg-blue-500/20 text-blue-300">
                                {permission.action}
                              </span>
                            </div>
                          </div>
                        </label>
                      );
                    })}
                  </div>
                </div>

                {/* Actions */}
                <div className="flex items-center justify-between p-6 border-t border-white/10 bg-white/5 flex-shrink-0">
                  <div className="text-sm text-gray-400">
                    {selectedPermissions.length} of {state.permissions.length}{" "}
                    permissions selected
                  </div>
                  <div className="flex items-center space-x-3">
                    <button
                      onClick={() =>
                        setState((prev) => ({
                          ...prev,
                          showRolePermissionModal: false,
                        }))
                      }
                      className={`${cosmic.button.secondary} px-6`}
                    >
                      Cancel
                    </button>
                    <button
                      onClick={handleAssignPermissions}
                      className={`${cosmic.button.primary} px-6 flex items-center space-x-2`}
                    >
                      <FaCheck className="w-4 h-4" />
                      <span>Update Permissions</span>
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </div>
    </AdminLayout>
  );
};

export default RoleManagement;
