import React, { useEffect, useState } from "react";
import AdminLayout from "./AdminLayout";
import { diamondApi, userApi } from "../services/api";
import type { Diamond, User } from "../types";
import { FaGem, FaChartBar, FaArrowUp, FaArrowDown } from "react-icons/fa";

interface DiamondTransfersState {
  transactions: Diamond[];
  loading: boolean;
  users: User[];
  filters: {
    userId: string;
    type: string;
    dateFrom: string;
    dateTo: string;
  };
  pagination: {
    page: number;
    limit: number;
    total: number;
  };
}

const DiamondTransfers: React.FC = () => {
  const [state, setState] = useState<DiamondTransfersState>({
    transactions: [],
    loading: true,
    users: [],
    filters: {
      userId: "",
      type: "",
      dateFrom: "",
      dateTo: "",
    },
    pagination: {
      page: 1,
      limit: 20,
      total: 0,
    },
  });

  useEffect(() => {
    fetchUsers();
    fetchTransactions();
  }, []);

  const fetchUsers = async () => {
    try {
      const response = await userApi.getUsers();
      setState((prev) => ({
        ...prev,
        users: response.users || [],
      }));
    } catch (error) {
      console.error("Failed to fetch users:", error);
    }
  };

  const fetchTransactions = async () => {
    setState((prev) => ({ ...prev, loading: true }));
    try {
      const response = await diamondApi.getAllTransactions(
        state.pagination.page,
        state.pagination.limit
      );
      setState((prev) => ({
        ...prev,
        transactions: response.transactions || [],
        pagination: {
          ...prev.pagination,
          total: response.total || 0,
        },
        loading: false,
      }));
    } catch (error) {
      console.error("Failed to fetch transactions:", error);
      setState((prev) => ({ ...prev, loading: false }));
    }
  };

  const handleFilterChange = (
    key: keyof typeof state.filters,
    value: string
  ) => {
    setState((prev) => ({
      ...prev,
      filters: {
        ...prev.filters,
        [key]: value,
      },
    }));
  };

  const applyFilters = () => {
    // In a real app, you'd send filters to the API
    // For now, we'll just filter the local data
    fetchTransactions();
  };

  const clearFilters = () => {
    setState((prev) => ({
      ...prev,
      filters: {
        userId: "",
        type: "",
        dateFrom: "",
        dateTo: "",
      },
    }));
    fetchTransactions();
  };

  const filteredTransactions = state.transactions.filter((transaction) => {
    const matchesUser =
      !state.filters.userId ||
      transaction.user_id.toString() === state.filters.userId;
    const matchesType =
      !state.filters.type || transaction.type === state.filters.type;
    const matchesDateFrom =
      !state.filters.dateFrom ||
      new Date(transaction.created_at) >= new Date(state.filters.dateFrom);
    const matchesDateTo =
      !state.filters.dateTo ||
      new Date(transaction.created_at) <= new Date(state.filters.dateTo);

    return matchesUser && matchesType && matchesDateFrom && matchesDateTo;
  });

  const totalAmount = filteredTransactions.reduce((sum, tx) => {
    return tx.type === "credit" ? sum + tx.amount : sum - tx.amount;
  }, 0);

  const creditTotal = filteredTransactions
    .filter((tx) => tx.type === "credit")
    .reduce((sum, tx) => sum + tx.amount, 0);

  const debitTotal = filteredTransactions
    .filter((tx) => tx.type === "debit")
    .reduce((sum, tx) => sum + tx.amount, 0);

  if (state.loading && state.transactions.length === 0) {
    return (
      <AdminLayout>
        <div className="flex justify-center items-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
            <p>Loading transactions...</p>
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
                Diamond Transfers
              </h1>
              <p className="mt-2 text-sm text-gray-700">
                View and manage all diamond transactions in the system.
              </p>
            </div>
            <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
              <button
                onClick={fetchTransactions}
                disabled={state.loading}
                className="inline-flex items-center justify-center rounded-md border border-transparent bg-indigo-600 px-4 py-2 text-sm font-medium text-white shadow-sm hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-indigo-500 focus:ring-offset-2 sm:w-auto disabled:opacity-50"
              >
                {state.loading ? "Loading..." : "Refresh"}
              </button>
            </div>
          </div>

          {/* Summary Cards */}
          <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-3">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-green-500 rounded-md flex items-center justify-center">
                      <FaArrowUp className="text-white text-sm" />
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Total Credits
                      </dt>
                      <dd className="text-lg font-medium text-gray-900 flex items-center space-x-1">
                        <span>+{creditTotal.toLocaleString()}</span>
                        <FaGem className="text-yellow-500" />
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-red-500 rounded-md flex items-center justify-center">
                      <FaArrowDown className="text-white text-sm" />
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Total Debits
                      </dt>
                      <dd className="text-lg font-medium text-gray-900 flex items-center space-x-1">
                        <span>-{debitTotal.toLocaleString()}</span>
                        <FaGem className="text-yellow-500" />
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-blue-500 rounded-md flex items-center justify-center">
                      <FaChartBar className="text-white text-sm" />
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-500 truncate">
                        Net Total
                      </dt>
                      <dd
                        className={`text-lg font-medium flex items-center space-x-1 ${
                          totalAmount >= 0 ? "text-green-600" : "text-red-600"
                        }`}
                      >
                        <span>
                          {totalAmount >= 0 ? "+" : ""}
                          {totalAmount.toLocaleString()}
                        </span>
                        <FaGem className="text-yellow-500" />
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Filters */}
          <div className="mt-8 bg-white shadow rounded-lg">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">Filters</h3>
            </div>
            <div className="p-6">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    User
                  </label>
                  <select
                    value={state.filters.userId}
                    onChange={(e) =>
                      handleFilterChange("userId", e.target.value)
                    }
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                  >
                    <option value="">All Users</option>
                    {state.users.map((user) => (
                      <option key={user.id} value={user.id}>
                        {user.username} ({user.email})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Type
                  </label>
                  <select
                    value={state.filters.type}
                    onChange={(e) => handleFilterChange("type", e.target.value)}
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                  >
                    <option value="">All Types</option>
                    <option value="credit">Credit</option>
                    <option value="debit">Debit</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    From Date
                  </label>
                  <input
                    type="date"
                    value={state.filters.dateFrom}
                    onChange={(e) =>
                      handleFilterChange("dateFrom", e.target.value)
                    }
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    To Date
                  </label>
                  <input
                    type="date"
                    value={state.filters.dateTo}
                    onChange={(e) =>
                      handleFilterChange("dateTo", e.target.value)
                    }
                    className="mt-1 block w-full rounded-md border-gray-300 shadow-sm focus:border-indigo-500 focus:ring-indigo-500 sm:text-sm"
                  />
                </div>
              </div>

              <div className="mt-4 flex space-x-3">
                <button
                  onClick={applyFilters}
                  className="inline-flex items-center px-4 py-2 border border-transparent text-sm font-medium rounded-md shadow-sm text-white bg-indigo-600 hover:bg-indigo-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Apply Filters
                </button>
                <button
                  onClick={clearFilters}
                  className="inline-flex items-center px-4 py-2 border border-gray-300 text-sm font-medium rounded-md shadow-sm text-gray-700 bg-white hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-indigo-500"
                >
                  Clear Filters
                </button>
              </div>
            </div>
          </div>

          {/* Transactions Table */}
          <div className="mt-8 flex flex-col">
            <div className="-my-2 -mx-4 overflow-x-auto sm:-mx-6 lg:-mx-8">
              <div className="inline-block min-w-full py-2 align-middle md:px-6 lg:px-8">
                <div className="overflow-hidden shadow ring-1 ring-black ring-opacity-5 md:rounded-lg">
                  <table className="min-w-full divide-y divide-gray-300">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Transaction ID
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          User
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Type
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Amount
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Balance After
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Description
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Date
                        </th>
                      </tr>
                    </thead>
                    <tbody className="bg-white divide-y divide-gray-200">
                      {filteredTransactions.map((transaction) => {
                        const user = state.users.find(
                          (u) => u.id === transaction.user_id
                        );
                        return (
                          <tr key={transaction.id}>
                            <td className="px-6 py-4 whitespace-nowrap text-sm font-mono text-gray-900">
                              {transaction.transaction_id}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <div className="text-sm font-medium text-gray-900">
                                {user?.username ||
                                  `User ${transaction.user_id}`}
                              </div>
                              <div className="text-sm text-gray-500">
                                {user?.email || `ID: ${transaction.user_id}`}
                              </div>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap">
                              <span
                                className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                  transaction.type === "credit"
                                    ? "bg-green-100 text-green-800"
                                    : "bg-red-100 text-red-800"
                                }`}
                              >
                                {transaction.type === "credit" ? (
                                  <span className="flex items-center space-x-1">
                                    <FaArrowUp />
                                    <span>Credit</span>
                                  </span>
                                ) : (
                                  <span className="flex items-center space-x-1">
                                    <FaArrowDown />
                                    <span>Debit</span>
                                  </span>
                                )}
                              </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              <span
                                className={`font-semibold ${
                                  transaction.type === "credit"
                                    ? "text-green-600"
                                    : "text-red-600"
                                }`}
                              >
                                <span className="flex items-center space-x-1">
                                  <span>
                                    {transaction.type === "credit" ? "+" : "-"}
                                    {transaction.amount.toLocaleString()}
                                  </span>
                                  <FaGem className="text-yellow-500" />
                                </span>
                              </span>
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                              <span className="flex items-center space-x-1">
                                <span>
                                  {transaction.balance.toLocaleString()}
                                </span>
                                <FaGem className="text-yellow-500" />
                              </span>
                            </td>
                            <td className="px-6 py-4 text-sm text-gray-900 max-w-xs truncate">
                              {transaction.description}
                            </td>
                            <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                              {new Date(
                                transaction.created_at
                              ).toLocaleDateString()}{" "}
                              {new Date(
                                transaction.created_at
                              ).toLocaleTimeString()}
                            </td>
                          </tr>
                        );
                      })}
                    </tbody>
                  </table>

                  {filteredTransactions.length === 0 && (
                    <div className="text-center py-12">
                      <FaGem className="text-4xl mb-4 mx-auto text-gray-400" />
                      <p className="text-gray-500">No transactions found</p>
                      <p className="text-sm text-gray-400 mt-2">
                        Try adjusting your filters or check back later
                      </p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
};

export default DiamondTransfers;
