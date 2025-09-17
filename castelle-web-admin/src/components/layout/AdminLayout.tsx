import React, { useState } from "react";
import { Link, useLocation } from "react-router-dom";
import { useAuth } from "../../contexts/AuthContext";
import {
  FaChartBar,
  FaGem,
  FaUsers,
  FaUserShield,
  FaSearch,
  FaBars,
  FaTimes,
  FaGamepad,
} from "react-icons/fa";
import { cosmic } from "../../styles/cosmic-theme";

interface AdminLayoutProps {
  children: React.ReactNode;
}

interface NavigationItem {
  name: string;
  href: string;
  icon: React.ReactElement;
  current: boolean;
}

const AdminLayout: React.FC<AdminLayoutProps> = ({ children }) => {
  const { user, logout } = useAuth();
  const location = useLocation();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  const navigation: NavigationItem[] = [
    {
      name: "Analytics",
      href: "/dashboard",
      icon: <FaChartBar className="text-lg" />,
      current: location.pathname === "/dashboard",
    },
    {
      name: "User Management",
      href: "/users",
      icon: <FaUsers className="text-lg" />,
      current: location.pathname === "/users",
    },
    {
      name: "Role Management",
      href: "/roles",
      icon: <FaUserShield className="text-lg" />,
      current: location.pathname === "/roles",
    },
    {
      name: "Diamond Transfers",
      href: "/diamonds",
      icon: <FaGem className="text-lg" />,
      current: location.pathname === "/diamonds",
    },
  ];

  return (
    <div className={`h-screen flex overflow-hidden ${cosmic.background}`}>
      {/* Mobile sidebar overlay */}
      {sidebarOpen && (
        <div className="fixed inset-0 flex z-40 md:hidden">
          <div
            className="fixed inset-0 bg-black/75 backdrop-blur-sm"
            onClick={() => setSidebarOpen(false)}
          />
          <div className="relative flex-1 flex flex-col max-w-xs w-full bg-gradient-to-b from-slate-900 to-purple-900 border-r border-white/10">
            <div className="absolute top-0 right-0 -mr-12 pt-2">
              <button
                className="ml-1 flex items-center justify-center h-10 w-10 rounded-full text-white hover:bg-white/10 focus:outline-none focus:ring-2 focus:ring-purple-500"
                onClick={() => setSidebarOpen(false)}
              >
                <FaTimes className="text-lg" />
              </button>
            </div>
            <SidebarContent navigation={navigation} />
          </div>
        </div>
      )}

      {/* Desktop sidebar */}
      <div className="hidden md:flex md:flex-shrink-0">
        <div className="flex flex-col w-64">
          <SidebarContent navigation={navigation} />
        </div>
      </div>

      {/* Main content */}
      <div className="flex flex-col flex-1 overflow-hidden">
        {/* Top navigation */}
        <div className="relative z-10 flex-shrink-0 flex h-16 bg-gradient-to-r from-slate-800/90 to-purple-800/90 backdrop-blur-md border-b border-white/10 shadow-lg">
          <button
            className="px-4 text-white hover:bg-white/10 focus:outline-none focus:ring-2 focus:ring-purple-500 md:hidden transition-colors"
            onClick={() => setSidebarOpen(true)}
          >
            <FaBars className="text-lg" />
          </button>
          <div className="flex-1 px-4 flex justify-between">
            <div className="flex-1 flex items-center max-w-lg">
              <div className="relative w-full">
                <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                  <FaSearch className="h-4 w-4 text-gray-400" />
                </div>
                <input
                  className={`${cosmic.input} w-full pl-10 pr-3 py-2`}
                  placeholder="Search admin panel..."
                  type="search"
                />
              </div>
            </div>
            <div className="ml-4 flex items-center md:ml-6">
              <div className="ml-3 relative">
                <div className="flex items-center space-x-4">
                  <div className="text-right">
                    <p className="text-sm text-white font-medium">
                      {user?.username}
                    </p>
                    <p className="text-xs text-gray-300">Administrator</p>
                  </div>
                  <button onClick={logout} className={cosmic.button.danger}>
                    Logout
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Page content */}
        <main className="flex-1 relative overflow-y-auto focus:outline-none bg-gradient-to-br from-slate-900 via-purple-900/20 to-slate-900">
          <div className="absolute inset-0 bg-[url('data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNjAiIGhlaWdodD0iNjAiIHZpZXdCb3g9IjAgMCA2MCA2MCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48ZGVmcz48cGF0dGVybiBpZD0iZ3JpZCIgd2lkdGg9IjYwIiBoZWlnaHQ9IjYwIiBwYXR0ZXJuVW5pdHM9InVzZXJTcGFjZU9uVXNlIj48cGF0aCBkPSJNIDYwIDAgTCAwIDAgMCA2MCIgZmlsbD0ibm9uZSIgc3Ryb2tlPSJyZ2JhKDI1NSwgMjU1LCAyNTUsIDAuMDUpIiBzdHJva2Utd2lkdGg9IjEiLz48L3BhdHRlcm4+PC9kZWZzPjxyZWN0IHdpZHRoPSIxMDAlIiBoZWlnaHQ9IjEwMCUiIGZpbGw9InVybCgjZ3JpZCkiLz48L3N2Zz4=')] opacity-50"></div>
          <div className="relative z-10">{children}</div>
        </main>
      </div>
    </div>
  );
};

interface SidebarContentProps {
  navigation: NavigationItem[];
}

const SidebarContent: React.FC<SidebarContentProps> = ({ navigation }) => {
  return (
    <div className="flex flex-col h-0 flex-1 bg-gradient-to-b from-slate-900 to-purple-900 border-r border-white/10">
      <div className="flex-1 flex flex-col pt-5 pb-4 overflow-y-auto">
        <div className="flex items-center flex-shrink-0 px-4 mb-8">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-lg bg-gradient-to-br from-purple-500 to-blue-500 shadow-lg">
              <FaGamepad className="text-white text-xl" />
            </div>
            <div>
              <h1 className="text-xl font-bold text-white">Caslette</h1>
              <p className="text-sm text-purple-300">Admin Portal</p>
            </div>
          </div>
        </div>
        <nav className="mt-5 flex-1 px-2 space-y-2">
          {navigation.map((item) => (
            <Link
              key={item.name}
              to={item.href}
              className={`group flex items-center px-3 py-2 text-sm font-medium rounded-lg transition-all duration-200 ${
                item.current
                  ? "bg-gradient-to-r from-purple-600/50 to-blue-600/50 text-white border border-purple-500/30 shadow-lg"
                  : "text-gray-300 hover:bg-white/10 hover:text-white hover:shadow-md"
              }`}
            >
              <span className="mr-3 opacity-75 group-hover:opacity-100 transition-opacity">
                {item.icon}
              </span>
              {item.name}
            </Link>
          ))}
        </nav>

        {/* Cosmic decoration */}
        <div className="mt-auto p-4">
          <div className="bg-gradient-to-r from-purple-500/10 to-blue-500/10 rounded-lg p-3 border border-purple-500/20">
            <div className="flex items-center space-x-2">
              <div className="w-2 h-2 bg-purple-400 rounded-full animate-pulse"></div>
              <div
                className="w-1 h-1 bg-blue-400 rounded-full animate-pulse"
                style={{ animationDelay: "0.5s" }}
              ></div>
              <div
                className="w-1 h-1 bg-purple-300 rounded-full animate-pulse"
                style={{ animationDelay: "1s" }}
              ></div>
            </div>
            <p className="text-xs text-gray-400 mt-2">Cosmic Admin v2.0</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AdminLayout;
