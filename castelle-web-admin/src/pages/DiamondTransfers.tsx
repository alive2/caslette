import React, { useEffect, useState } from "react";
import AdminLayout from "../components/layout/AdminLayout";
import DataTable from "../components/common/DataTable";
import type {
  DataTableColumn,
  DataTableAction,
} from "../components/common/DataTable";
import { diamondApi, userApi } from "../services/api";
import type { Diamond, User } from "../types";
import { FaGem, FaChartBar, FaArrowUp, FaArrowDown } from "react-icons/fa";
import { cosmic } from "../styles/cosmic-theme";

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

  // Helper function to find user
  const getUserForTransaction = (transaction: Diamond) => {
    return state.users.find((u) => u.id === transaction.user_id);
  };

  // Define DataTable columns
  const transactionColumns: DataTableColumn<Diamond>[] = [
    {
      key: "transaction_id",
      title: "Transaction ID",
      sortable: true,
      className: "font-mono",
      mobileLabel: "Transaction ID",
      hiddenOnMobile: true,
    },
    {
      key: "user_id",
      title: "User",
      sortable: true,
      render: (_, transaction) => {
        const user = getUserForTransaction(transaction);
        return (
          <div>
            <div className="text-sm font-medium text-white">
              {user?.username || `User ${transaction.user_id}`}
            </div>
            <div className="text-sm text-gray-400">
              {user?.email || `ID: ${transaction.user_id}`}
            </div>
          </div>
        );
      },
      mobileLabel: "User",
    },
    {
      key: "type",
      title: "Type",
      sortable: true,
      render: (_, transaction) => (
        <span
          className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
            transaction.type === "credit"
              ? "bg-green-500/20 text-green-400 border border-green-500/30"
              : "bg-red-500/20 text-red-400 border border-red-500/30"
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
      ),
      mobileLabel: "Type",
    },
    {
      key: "amount",
      title: "Amount",
      sortable: true,
      render: (_, transaction) => (
        <span
          className={`font-semibold ${
            transaction.type === "credit" ? "text-green-400" : "text-red-400"
          }`}
        >
          <span className="flex items-center justify-end space-x-1">
            <span>
              {transaction.type === "credit" ? "+" : "-"}
              {Math.abs(transaction.amount).toLocaleString()}
            </span>
            <FaGem className="text-yellow-400" />
          </span>
        </span>
      ),
      mobileLabel: "Amount",
    },
    {
      key: "balance",
      title: "Balance After",
      sortable: true,
      render: (_, transaction) => (
        <span className="flex items-center justify-end space-x-1">
          <span>{transaction.balance.toLocaleString()}</span>
          <FaGem className="text-yellow-400" />
        </span>
      ),
      mobileLabel: "Balance After",
    },
    {
      key: "description",
      title: "Description",
      render: (_, transaction) => (
        <span className="max-w-xs truncate block">
          {transaction.description}
        </span>
      ),
      mobileLabel: "Description",
    },
    {
      key: "created_at",
      title: "Date",
      sortable: true,
      render: (_, transaction) => (
        <div>
          <div className="text-sm text-white">
            {new Date(transaction.created_at).toLocaleDateString()}
          </div>
          <div className="text-xs text-gray-400">
            {new Date(transaction.created_at).toLocaleTimeString()}
          </div>
        </div>
      ),
      mobileLabel: "Date",
    },
  ];

  // Define DataTable actions (none for transactions, but could add view details)
  const transactionActions: DataTableAction<Diamond>[] = [];

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
    .reduce((sum, tx) => sum + Math.abs(tx.amount), 0);

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
      <div className="p-6">
        <div className="max-w-7xl mx-auto">
          <div className="sm:flex sm:items-center mb-8">
            <div className="sm:flex-auto">
              <h1 className="text-3xl font-bold text-white flex items-center space-x-3">
                <FaGem className="text-purple-400" />
                <span>Diamond Transfers</span>
              </h1>
              <p className="mt-2 text-gray-300">
                View and manage all diamond transactions in the system.
              </p>
            </div>
            <div className="mt-4 sm:mt-0 sm:ml-16 sm:flex-none">
              <button
                onClick={fetchTransactions}
                disabled={state.loading}
                className={`${cosmic.button.primary} ${
                  state.loading ? "opacity-50 cursor-not-allowed" : ""
                }`}
              >
                {state.loading ? "Loading..." : "Refresh"}
              </button>
            </div>
          </div>

          {/* Summary Cards */}
          <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-3">
            <div className={cosmic.cardElevated}>
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-green-500/20 border border-green-500/30 rounded-md flex items-center justify-center">
                      <FaArrowUp className="text-green-400 text-sm" />
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-400 truncate">
                        Total Credits
                      </dt>
                      <dd className="text-lg font-medium text-white flex items-center space-x-1">
                        <span>+{creditTotal.toLocaleString()}</span>
                        <FaGem className="text-yellow-400" />
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className={cosmic.cardElevated}>
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-red-500/20 border border-red-500/30 rounded-md flex items-center justify-center">
                      <FaArrowDown className="text-red-400 text-sm" />
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-400 truncate">
                        Total Debits
                      </dt>
                      <dd className="text-lg font-medium text-white flex items-center space-x-1">
                        <span>-{debitTotal.toLocaleString()}</span>
                        <FaGem className="text-yellow-400" />
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>

            <div className={cosmic.cardElevated}>
              <div className="p-5">
                <div className="flex items-center">
                  <div className="flex-shrink-0">
                    <div className="w-8 h-8 bg-blue-500/20 border border-blue-500/30 rounded-md flex items-center justify-center">
                      <FaChartBar className="text-blue-400 text-sm" />
                    </div>
                  </div>
                  <div className="ml-5 w-0 flex-1">
                    <dl>
                      <dt className="text-sm font-medium text-gray-400 truncate">
                        Net Total
                      </dt>
                      <dd
                        className={`text-lg font-medium flex items-center space-x-1 ${
                          totalAmount >= 0 ? "text-green-400" : "text-red-400"
                        }`}
                      >
                        <span>
                          {totalAmount >= 0 ? "+" : ""}
                          {totalAmount.toLocaleString()}
                        </span>
                        <FaGem className="text-yellow-400" />
                      </dd>
                    </dl>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Filters */}
          <div className={`mt-8 ${cosmic.cardElevated}`}>
            <div className="px-6 py-4 border-b border-white/10">
              <h3 className="text-lg font-medium text-white">Filters</h3>
            </div>
            <div className="p-6">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
                <div>
                  <label className="block text-sm font-medium text-gray-300">
                    User
                  </label>
                  <select
                    value={state.filters.userId}
                    onChange={(e) =>
                      handleFilterChange("userId", e.target.value)
                    }
                    className={cosmic.select}
                    style={{
                      backgroundColor: "rgba(255, 255, 255, 0.1)",
                      color: "white",
                    }}
                  >
                    <option
                      value=""
                      style={{ backgroundColor: "#1f2937", color: "white" }}
                    >
                      All Users
                    </option>
                    {state.users.map((user) => (
                      <option
                        key={user.id}
                        value={user.id}
                        style={{ backgroundColor: "#1f2937", color: "white" }}
                      >
                        {user.username} ({user.email})
                      </option>
                    ))}
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300">
                    Type
                  </label>
                  <select
                    value={state.filters.type}
                    onChange={(e) => handleFilterChange("type", e.target.value)}
                    className={cosmic.select}
                    style={{
                      backgroundColor: "rgba(255, 255, 255, 0.1)",
                      color: "white",
                    }}
                  >
                    <option
                      value=""
                      style={{ backgroundColor: "#1f2937", color: "white" }}
                    >
                      All Types
                    </option>
                    <option
                      value="credit"
                      style={{ backgroundColor: "#1f2937", color: "white" }}
                    >
                      Credit
                    </option>
                    <option
                      value="debit"
                      style={{ backgroundColor: "#1f2937", color: "white" }}
                    >
                      Debit
                    </option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300">
                    From Date
                  </label>
                  <input
                    type="date"
                    value={state.filters.dateFrom}
                    onChange={(e) =>
                      handleFilterChange("dateFrom", e.target.value)
                    }
                    className={cosmic.input}
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-300">
                    To Date
                  </label>
                  <input
                    type="date"
                    value={state.filters.dateTo}
                    onChange={(e) =>
                      handleFilterChange("dateTo", e.target.value)
                    }
                    className={cosmic.input}
                  />
                </div>
              </div>

              <div className="mt-4 flex space-x-3">
                <button
                  onClick={applyFilters}
                  className={cosmic.button.primary}
                >
                  Apply Filters
                </button>
                <button
                  onClick={clearFilters}
                  className={cosmic.button.secondary}
                >
                  Clear Filters
                </button>
              </div>
            </div>
          </div>

          {/* Transactions Table */}
          <DataTable
            data={filteredTransactions}
            columns={transactionColumns}
            actions={transactionActions}
            loading={state.loading && state.transactions.length === 0}
            searchable={false} // We have custom filters above
            emptyMessage="No transactions found. Try adjusting your filters or check back later."
            cardTitleExtractor={(transaction) => {
              return `${transaction.type === "credit" ? "+" : "-"}${Math.abs(
                transaction.amount
              ).toLocaleString()} diamonds`;
            }}
            cardSubtitleExtractor={(transaction) => {
              const user = getUserForTransaction(transaction);
              return `${
                user?.username || `User ${transaction.user_id}`
              } • ${new Date(transaction.created_at).toLocaleDateString()}`;
            }}
            cardKeyExtractor={(transaction) => transaction.id.toString()}
            className="mt-8"
          />
        </div>
      </div>
    </AdminLayout>
  );
};

export default DiamondTransfers;
