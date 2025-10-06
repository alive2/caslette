import React, { useEffect, useState } from "react";
import AdminLayout from "../components/layout/AdminLayout";
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
import { cosmic } from "../styles/cosmic-theme";

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
        const users = usersResponse.data?.users || [];

        // Fetch all diamond transactions
        const diamondsResponse = await diamondApi.getAllTransactions();
        const transactions = diamondsResponse.data?.transactions || [];

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
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-purple-500 mx-auto mb-4"></div>
            <p className="text-white">Loading analytics...</p>
          </div>
        </div>
      </AdminLayout>
    );
  }

  const stats = [
    {
      name: "Total Users",
      stat: analyticsData?.totalUsers || 0,
      icon: <FaUsers className="text-blue-400" />,
      color: "bg-blue-500/20 border-blue-500/30",
    },
    {
      name: "Active Users",
      stat: analyticsData?.activeUsers || 0,
      icon: <FaCheck className="text-green-400" />,
      color: "bg-green-500/20 border-green-500/30",
    },
    {
      name: "Total Diamonds",
      stat: (analyticsData?.totalDiamonds || 0).toLocaleString(),
      icon: <FaGem className="text-yellow-400" />,
      color: "bg-yellow-500/20 border-yellow-500/30",
    },
    {
      name: "Today's Transactions",
      stat: analyticsData?.todayTransactions || 0,
      icon: <FaChartBar className="text-purple-400" />,
      color: "bg-purple-500/20 border-purple-500/30",
    },
  ];

  return (
    <AdminLayout>
      <div className="p-6">
        <div className="max-w-7xl mx-auto">
          <h1 className="text-3xl font-bold text-white flex items-center space-x-3 mb-8">
            <FaChartLine className="text-purple-400" />
            <span>Analytics Dashboard</span>
          </h1>

          {/* Stats Grid */}
          <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 lg:grid-cols-4 mb-8">
            {stats.map((item) => (
              <div key={item.name} className={cosmic.cardElevated}>
                <div className="p-5">
                  <div className="flex items-center">
                    <div className="flex-shrink-0">
                      <div
                        className={`w-8 h-8 ${item.color} border rounded-md flex items-center justify-center`}
                      >
                        {item.icon}
                      </div>
                    </div>
                    <div className="ml-5 w-0 flex-1">
                      <dl>
                        <dt className="text-sm font-medium text-gray-300 truncate">
                          {item.name}
                        </dt>
                        <dd className="text-lg font-medium text-white">
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
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
            {/* Recent Users */}
            <div className={cosmic.cardElevated}>
              <div className="p-6">
                <h3 className="text-lg leading-6 font-medium text-white mb-4 flex items-center space-x-2">
                  <FaUsers className="text-blue-400" />
                  <span>Recent Users</span>
                </h3>
                <div className="flow-root">
                  <ul className="-my-5 divide-y divide-white/10">
                    {analyticsData?.recentUsers.map((user: any) => (
                      <li key={user.id} className="py-4">
                        <div className="flex items-center space-x-4">
                          <div className="flex-shrink-0">
                            <div className="h-8 w-8 bg-purple-500/20 border border-purple-500/30 rounded-full flex items-center justify-center">
                              <span className="text-sm font-medium text-purple-300">
                                {user.username.charAt(0).toUpperCase()}
                              </span>
                            </div>
                          </div>
                          <div className="flex-1 min-w-0">
                            <p className="text-sm font-medium text-white truncate">
                              {user.username}
                            </p>
                            <p className="text-sm text-gray-400 truncate">
                              {user.email}
                            </p>
                          </div>
                          <div className="flex-shrink-0">
                            <span
                              className={`inline-flex px-2 py-1 text-xs font-semibold rounded-full ${
                                user.is_active
                                  ? "bg-green-500/20 text-green-400 border border-green-500/30"
                                  : "bg-red-500/20 text-red-400 border border-red-500/30"
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
            <div className={cosmic.cardElevated}>
              <div className="p-6">
                <h3 className="text-lg leading-6 font-medium text-white mb-4 flex items-center space-x-2">
                  <FaGem className="text-yellow-400" />
                  <span>Recent Transactions</span>
                </h3>
                <div className="flow-root">
                  <ul className="-my-5 divide-y divide-white/10">
                    {analyticsData?.recentTransactions.map(
                      (transaction: any) => (
                        <li key={transaction.id} className="py-4">
                          <div className="flex items-center space-x-4">
                            <div className="flex-shrink-0">
                              <div
                                className={`w-8 h-8 rounded-md flex items-center justify-center ${
                                  transaction.type === "credit"
                                    ? "bg-green-500/20 border border-green-500/30"
                                    : "bg-red-500/20 border border-red-500/30"
                                }`}
                              >
                                {transaction.type === "credit" ? (
                                  <FaArrowUp className="text-green-400" />
                                ) : (
                                  <FaArrowDown className="text-red-400" />
                                )}
                              </div>
                            </div>
                            <div className="flex-1 min-w-0">
                              <p className="text-sm font-medium text-white truncate">
                                {transaction.description}
                              </p>
                              <p className="text-sm text-gray-400 truncate">
                                User ID: {transaction.user_id}
                              </p>
                            </div>
                            <div className="flex-shrink-0">
                              <span
                                className={`inline-flex items-center px-2 py-1 text-xs font-semibold rounded-full space-x-1 ${
                                  transaction.type === "credit"
                                    ? "bg-green-500/20 text-green-400 border border-green-500/30"
                                    : "bg-red-500/20 text-red-400 border border-red-500/30"
                                }`}
                              >
                                <span>
                                  {transaction.type === "credit" ? "+" : "-"}
                                  {Math.abs(transaction.amount)}
                                </span>
                                <FaGem className="text-yellow-400" />
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
            <div className={cosmic.cardElevated}>
              <div className="p-6">
                <h3 className="text-lg leading-6 font-medium text-white mb-4 flex items-center space-x-2">
                  <FaChartLine className="text-blue-400" />
                  <span>Activity Overview</span>
                </h3>
                <div className="h-64 bg-white/5 border border-white/10 rounded-lg flex items-center justify-center">
                  <div className="text-center">
                    <FaChartLine className="text-4xl mb-2 mx-auto text-gray-400" />
                    <p className="text-gray-300">
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
