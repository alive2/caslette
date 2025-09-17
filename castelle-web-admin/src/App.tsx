import React from "react";
import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import AdminLogin from "./pages/AdminLogin";
import AdminAnalytics from "./pages/AdminAnalytics";
import UserManagement from "./pages/UserManagement";
import RoleManagement from "./pages/RoleManagement";
import DiamondTransfers from "./pages/DiamondTransfers";

const ProtectedRoute: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const { user, loading, isAdmin } = useAuth();

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        Loading...
      </div>
    );
  }

  return user && isAdmin ? <>{children}</> : <Navigate to="/login" />;
};

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Router>
        <div className="App">
          <Routes>
            <Route path="/login" element={<AdminLogin />} />
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <AdminAnalytics />
                </ProtectedRoute>
              }
            />
            <Route
              path="/users"
              element={
                <ProtectedRoute>
                  <UserManagement />
                </ProtectedRoute>
              }
            />
            <Route
              path="/roles"
              element={
                <ProtectedRoute>
                  <RoleManagement />
                </ProtectedRoute>
              }
            />
            <Route
              path="/diamonds"
              element={
                <ProtectedRoute>
                  <DiamondTransfers />
                </ProtectedRoute>
              }
            />
            <Route path="/" element={<Navigate to="/dashboard" />} />
          </Routes>
        </div>
      </Router>
    </AuthProvider>
  );
};

export default App;
