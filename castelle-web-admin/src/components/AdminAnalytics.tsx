import React, { useEffect, useState } from "react";
import AdminLayout from "./AdminLayout";
import { userApi, diamondApi } from "../services/api";
import {
  FaGem,
  FaChartBar,
  FaArrowUp,
  FaArrowDown,
  FaChartLine,
  FaUsers,
  FaCheck,
} from "react-icons/fa";

interface AnalyticsData {
  totalUsers: number;
  activeUsers: number;
  totalDiamonds: number;
  todayTransactions: number;
  recentUsers: any[];
  recentTransactions: any[];
}

const AdminAnalytics: React.FC = () => {
  const [analyticsData, setAnalyticsData] = useState<AnalyticsData | null>(
    null
  );
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchAnalytics = async () => {
      try {
        // Fetch users
        const usersResponse = await userApi.getUsers();
        const users = usersResponse.users || [];

        // Fetch all diamond transactions
        const diamondsResponse = await diamondApi.getAllTransactions();
        const transactions = diamondsResponse.transactions || [];

        // Calculate analytics
        const totalUsers = users.length;
        const activeUsers = users.filter((user: any) => user.is_active).length;

        const totalDiamonds = transactions.reduce((sum: number, tx: any) => {
          return tx.type === "credit" ? sum + tx.amount : sum - tx.amount;
        }, 0);

        const today = new Date().toISOString().split("T")[0];
        const todayTransactions = transactions.filter((tx: any) =>
          tx.created_at.startsWith(today)
        ).length;

        setAnalyticsData({
          totalUsers,
          activeUsers,
          totalDiamonds,
          todayTransactions,
          recentUsers: users.slice(-5).reverse(),
          recentTransactions: transactions.slice(-10).reverse(),
        });
      } catch (error) {
        console.error("Failed to fetch analytics:", error);
      } finally {
        setLoading(false);
      }
    };

    fetchAnalytics();
  }, []);

  if (loading) {
    return (
      <AdminLayout>
        <div className="flex justify-center items-center h-full">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
            <p>Loading analytics...</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  const stats = [
    {
      name: "Total Users",
      stat: analyticsData?.totalUsers || 0,
      icon: <FaUsers className="text-white" />,
      color: "bg-blue-500",
    },
    {
      name: "Active Users",
      stat: analyticsData?.activeUsers || 0,
      icon: <FaCheck className="text-white" />,
      color: "bg-green-500",
    },
    {
      name: "Total Diamonds",
      stat: (analyticsData?.totalDiamonds || 0).toLocaleString(),
      icon: <FaGem className="text-white" />,
      color: "bg-yellow-500",
    },
    {
      name: "Today's Transactions",
      stat: analyticsData?.todayTransactions || 0,
      icon: <FaChartBar className="text-white" />,
      color: "bg-purple-500",
    },
  ];

  return (
    <AdminLayout>
      <div className="py-6">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
          <h1 className="text-2xl font-semibold text-gray-900">
            Analytics Dashboard
          </h1>
        </div>
        <div className="max-w-7xl mx-auto px-4 sm:px-6 md:px-8">
          {/* Stats Grid */}
          <div className="mt-6 grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4">
            {stats.map((item) => (
              <div
                key={item.name}
                className="bg-white overflow-hidden shadow rounded-lg"
              >
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <div
                        className={`w-8 h-8 ${item.color} rounded-md flex items-center justify-center`}
                      >
                        <span className="text-white text-sm">{item.icon}</span>
                      </div>
                    </div>
                    <div className="ml-5 w-0 flex-1">
                      <dl>
                        <dt className="text-sm font-medium text-gray-500 truncate">
                          {item.name}
                        </dt>
                        <dd className="text-lg font-medium text-gray-900">
                          {item.stat}
                        </dd>
                      </dl>
                    </div>
                  </div>
                </div>
              </div>
            ))}
          </div>

          {/* Charts Section */}
          <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Recent Users */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Recent Users
                </h3>
                <div className="flow-root">
                  <ul className="-my-5 divide-y divide-gray-200">
                    {analyticsData?.recentUsers.map((user: any) => (
                      <li key={user.id} className="py-4">
                        <div className="flex items-center space-x-4">
                          <div className="flex-shrink-0">
                            <div className="h-8 w-8 bg-gray-300 rounded-full flex items-center justify-center">
                              <span className="text-sm font-medium text-gray-700">
                                {user.username.charAt(0).toUpperCase()}
                              </span>
                            </div>
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-gray-900 truncate">
                              {user.username}
                            </p>
                            <p className="text-sm text-gray-500 truncate">
                              {user.email}
                            </p>
                          </div>
                          <div className="flex-shrink-0">
                            <span
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                user.is_active
                                  ? "bg-green-100 text-green-800"
                                  : "bg-red-100 text-red-800"
                              }`}
                            >
                              {user.is_active ? "Active" : "Inactive"}
                            </span>
                          </div>
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            </div>

            {/* Recent Transactions */}
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Recent Transactions
                </h3>
                <div className="flow-root">
                  <ul className="-my-5 divide-y divide-gray-200">
                    {analyticsData?.recentTransactions.map(
                      (transaction: any) => (
                        <li key={transaction.id} className="py-4">
                          <div className="flex items-center space-x-4">
                            <div className="flex-shrink-0">
                              <span className="text-lg text-gray-600">
                                {transaction.type === "credit" ? (
                                  <FaArrowUp className="text-green-600" />
                                ) : (
                                  <FaArrowDown className="text-red-600" />
                                )}
                              </span>
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-medium text-gray-900 truncate">
                                {transaction.description}
                              </p>
                              <p className="text-sm text-gray-500 truncate">
                                User ID: {transaction.user_id}
                              </p>
                            </div>
                            <div className="flex-shrink-0">
                              <span
                                className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                  transaction.type === "credit"
                                    ? "bg-green-100 text-green-800"
                                    : "bg-red-100 text-red-800"
                                }`}
                              >
                                {transaction.type === "credit" ? "+" : "-"}
                                {transaction.amount}{" "}
                                <FaGem className="inline text-yellow-500" />
                              </span>
                            </div>
                          </div>
                        </li>
                      )
                    )}
                  </ul>
                </div>
              </div>
            </div>
          </div>

          {/* Activity Chart Placeholder */}
          <div className="mt-8">
            <div className="bg-white overflow-hidden shadow rounded-lg">
              <div className="p-6">
                <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                  Activity Overview
                </h3>
                <div className="h-64 bg-gray-50 rounded-lg flex items-center justify-center">
                  <div className="text-center">
                    <FaChartLine className="text-4xl mb-2 mx-auto text-gray-400" />
                    <p className="text-gray-500">
                      Activity chart coming soon...
                    </p>
                    <p className="text-sm text-gray-400 mt-2">
                      This will show user activity and transaction trends over
                      time
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AdminLayout>
  );
};

export default AdminAnalytics;
