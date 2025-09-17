import React, { useState, useEffect } from "react";
import { useAuth } from "../contexts/AuthContext";
import { useNavigate } from "react-router-dom";
import { cosmic } from "../styles/cosmic-theme";
import { FaUser, FaLock, FaSignInAlt, FaGamepad } from "react-icons/fa";

const AdminLogin: React.FC = () => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const { login, user, isAdmin } = useAuth();
  const navigate = useNavigate();

  // Redirect to dashboard if user is already authenticated and is an admin
  useEffect(() => {
    if (user && isAdmin) {
      navigate("/dashboard");
    }
  }, [user, isAdmin, navigate]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError("");

    try {
      await login(username, password);
    } catch {
      setError("Invalid credentials or insufficient privileges");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      className={`min-h-screen flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8 ${cosmic.background}`}
    >
      {/* Cosmic Background Elements */}
      <div className="absolute inset-0 overflow-hidden">
        <div className="absolute top-1/4 left-1/4 w-64 h-64 bg-purple-500/20 rounded-full blur-3xl"></div>
        <div className="absolute bottom-1/4 right-1/4 w-80 h-80 bg-blue-500/20 rounded-full blur-3xl"></div>
        <div className="absolute top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2 w-96 h-96 bg-pink-500/10 rounded-full blur-3xl"></div>
      </div>

      <div className="relative z-10 max-w-md w-full space-y-8">
        <div className="text-center">
          <div className="flex justify-center mb-6">
            <div className="w-20 h-20 bg-gradient-to-br from-purple-500 to-blue-600 rounded-2xl flex items-center justify-center shadow-xl">
              <FaGamepad className="text-white text-3xl" />
            </div>
          </div>
          <h2 className="text-4xl font-bold text-white mb-2">Admin Portal</h2>
          <p className="text-gray-300">
            Sign in with your administrator account
          </p>
        </div>

        <div className={cosmic.cardElevated}>
          <div className="p-8">
            <form className="space-y-6" onSubmit={handleSubmit}>
              <div className="space-y-4">
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <FaUser className="text-gray-400" />
                  </div>
                  <input
                    type="text"
                    required
                    className={`${cosmic.input} pl-10`}
                    placeholder="Username"
                    value={username}
                    onChange={(e) => setUsername(e.target.value)}
                  />
                </div>
                <div className="relative">
                  <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                    <FaLock className="text-gray-400" />
                  </div>
                  <input
                    type="password"
                    required
                    className={`${cosmic.input} pl-10`}
                    placeholder="Password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                  />
                </div>
              </div>

              {error && (
                <div className="text-red-400 text-sm text-center bg-red-500/10 border border-red-500/20 rounded-lg p-3">
                  {error}
                </div>
              )}

              <button
                type="submit"
                disabled={loading}
                className={`${
                  cosmic.button.primary
                } w-full flex justify-center items-center space-x-2 ${
                  loading ? "opacity-50 cursor-not-allowed" : ""
                }`}
              >
                <FaSignInAlt />
                <span>{loading ? "Signing in..." : "Sign in"}</span>
              </button>
            </form>
          </div>
        </div>

        <div className="text-center">
          <p className="text-gray-400 text-sm">
            Restricted area for authorized personnel only
          </p>
        </div>
      </div>
    </div>
  );
};

export default AdminLogin;
