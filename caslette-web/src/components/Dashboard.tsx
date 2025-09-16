import React, { useEffect, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { useWebSocket } from "../contexts/WebSocketContext";
import { diamondApi } from "../services/api";
import type { Diamond } from "../types";

const Dashboard: React.FC = () => {
  const { user, logout } = useAuth();
  const { isConnected, sendMessage, onMessage, offMessage } = useWebSocket();
  const [diamonds, setDiamonds] = useState<Diamond[]>([]);
  const [currentBalance, setCurrentBalance] = useState(0);
  const [loading, setLoading] = useState(true);
  const [notifications, setNotifications] = useState<string[]>([]);

  useEffect(() => {
    const fetchDiamonds = async () => {
      if (user) {
        try {
          const response = await diamondApi.getUserDiamonds(user.id);
          setDiamonds(response.diamonds);
          setCurrentBalance(response.current_balance);
        } catch (error) {
          console.error("Failed to fetch diamonds:", error);
        } finally {
          setLoading(false);
        }
      }
    };

    fetchDiamonds();

    // Set up WebSocket listeners for real-time updates
    const handleBalanceUpdate = (data: Record<string, unknown>) => {
      if (typeof data.balance === "number") {
        setCurrentBalance(data.balance);
      }
      if (data.diamonds) {
        // Refresh diamonds list
        fetchDiamonds();
      }
    };

    const handleNotification = (data: Record<string, unknown>) => {
      if (typeof data.message === "string") {
        setNotifications((prev) => [...prev, data.message as string]);
        // Auto-remove notification after 5 seconds
        setTimeout(() => {
          setNotifications((prev) => prev.filter((n) => n !== data.message));
        }, 5000);
      }
    };

    onMessage("balance_update", handleBalanceUpdate);
    onMessage("notification", handleNotification);

    return () => {
      offMessage("balance_update", handleBalanceUpdate);
      offMessage("notification", handleNotification);
    };
  }, [user, onMessage, offMessage]);

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        Loading...
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="bg-white shadow">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-6">
            <div className="flex items-center">
              <h1 className="text-2xl font-bold text-gray-900">
                Caslette Dashboard
              </h1>
            </div>
            <div className="flex items-center space-x-4">
              <div className="flex items-center space-x-2">
                <span
                  className={`w-2 h-2 rounded-full ${
                    isConnected ? "bg-green-500" : "bg-red-500"
                  }`}
                />
                <span className="text-sm text-gray-500">
                  {isConnected ? "Connected" : "Disconnected"}
                </span>
              </div>
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

      <div className="max-w-7xl mx-auto py-6 sm:px-6 lg:px-8">
        <div className="px-4 py-6 sm:px-0">
          {/* Notifications */}
          {notifications.length > 0 && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
              <h3 className="text-sm font-medium text-blue-800 mb-2">
                Notifications
              </h3>
              <div className="space-y-1">
                {notifications.map((notification, index) => (
                  <div key={index} className="text-sm text-blue-700">
                    {notification}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* WebSocket Test Button */}
          <div className="bg-white overflow-hidden shadow rounded-lg mb-6">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Real-time Features
              </h3>
              <button
                onClick={() =>
                  sendMessage("test", { message: "Hello from frontend!" })
                }
                disabled={!isConnected}
                className={`px-4 py-2 rounded-md text-sm font-medium ${
                  isConnected
                    ? "bg-blue-600 hover:bg-blue-700 text-white"
                    : "bg-gray-300 text-gray-500 cursor-not-allowed"
                }`}
              >
                Send Test Message
              </button>
            </div>
          </div>

          {/* Diamond Balance Card */}
          <div className="bg-white overflow-hidden shadow rounded-lg mb-6">
            <div className="px-4 py-5 sm:p-6">
              <div className="flex items-center">
                <div className="flex-shrink-0">
                  <div className="w-8 h-8 bg-yellow-400 rounded-full flex items-center justify-center">
                    <span className="text-yellow-800 font-bold">💎</span>
                  </div>
                </div>
                <div className="ml-5 w-0 flex-1">
                  <dl>
                    <dt className="text-sm font-medium text-gray-500 truncate">
                      Diamond Balance
                    </dt>
                    <dd className="text-lg font-medium text-gray-900">
                      {currentBalance.toLocaleString()} diamonds
                    </dd>
                  </dl>
                </div>
              </div>
            </div>
          </div>

          {/* User Info Card */}
          <div className="bg-white overflow-hidden shadow rounded-lg mb-6">
            <div className="px-4 py-5 sm:p-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900 mb-4">
                Profile Information
              </h3>
              <div className="grid grid-cols-1 gap-x-4 gap-y-6 sm:grid-cols-2">
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Username
                  </label>
                  <div className="mt-1 text-sm text-gray-900">
                    {user?.username}
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Email
                  </label>
                  <div className="mt-1 text-sm text-gray-900">
                    {user?.email}
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    First Name
                  </label>
                  <div className="mt-1 text-sm text-gray-900">
                    {user?.first_name || "Not provided"}
                  </div>
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700">
                    Last Name
                  </label>
                  <div className="mt-1 text-sm text-gray-900">
                    {user?.last_name || "Not provided"}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Transaction History */}
          <div className="bg-white shadow overflow-hidden sm:rounded-md">
            <div className="px-4 py-5 sm:px-6">
              <h3 className="text-lg leading-6 font-medium text-gray-900">
                Recent Diamond Transactions
              </h3>
              <p className="mt-1 max-w-2xl text-sm text-gray-500">
                Your latest diamond transaction history
              </p>
            </div>
            {diamonds.length > 0 ? (
              <ul className="divide-y divide-gray-200">
                {diamonds.slice(0, 10).map((diamond) => (
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
                            {diamond.description ||
                              `${diamond.type} transaction`}
                          </div>
                          <div className="text-sm text-gray-500">
                            {new Date(diamond.created_at).toLocaleDateString()}
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
            ) : (
              <div className="px-4 py-6 text-center text-gray-500">
                No diamond transactions yet
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default Dashboard;
