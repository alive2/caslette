import React, { useEffect, useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { useWebSocket } from "../contexts/WebSocketContext";
import { diamondApi } from "../services/api";
import type { Diamond } from "../types";
import { FaGem, FaDice, FaTrophy, FaBullseye, FaStar } from "react-icons/fa";
import "../styles/cosmic.css";

const HomePage: React.FC = () => {
  const { user, logout } = useAuth();
  const { isConnected, sendMessage, onMessage, offMessage } = useWebSocket();
  const [diamonds, setDiamonds] = useState<Diamond[]>([]);
  const [currentBalance, setCurrentBalance] = useState(0);
  const [loading, setLoading] = useState(true);
  const [gameNotifications, setGameNotifications] = useState<string[]>([]);
  const [gameState, setGameState] = useState<"idle" | "playing" | "finished">(
    "idle"
  );

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

    // WebSocket message handlers
    const handleBalanceUpdate = (data: Record<string, unknown>) => {
      const balance = data.balance as number;
      const diamonds = data.diamonds as Diamond[];
      setCurrentBalance(balance);
      if (diamonds) {
        setDiamonds(diamonds);
      }
      setGameNotifications((prev) => [
        ...prev,
        `Balance updated: ${balance} diamonds`,
      ]);
    };

    const handleGameUpdate = (data: Record<string, unknown>) => {
      const state = data.state as string;
      const message = data.message as string | undefined;
      setGameState(state as "idle" | "playing" | "finished");
      if (message) {
        setGameNotifications((prev) => [...prev, message]);
      }
    };

    const handleNotification = (data: Record<string, unknown>) => {
      const message = data.message as string;
      if (message) {
        setGameNotifications((prev) => [...prev, message]);
      }
    };

    // Register WebSocket listeners
    onMessage("balance_update", handleBalanceUpdate);
    onMessage("game_update", handleGameUpdate);
    onMessage("notification", handleNotification);

    return () => {
      offMessage("balance_update", handleBalanceUpdate);
      offMessage("game_update", handleGameUpdate);
      offMessage("notification", handleNotification);
    };
  }, [user, onMessage, offMessage]);

  const startGame = () => {
    if (currentBalance >= 10) {
      setGameState("playing");
      sendMessage("start_game", { bet_amount: 10 });
      setGameNotifications((prev) => [
        ...prev,
        "Game started with 10 diamonds bet",
      ]);
    } else {
      setGameNotifications((prev) => [
        ...prev,
        "Insufficient diamonds to start game",
      ]);
    }
  };

  const simulateWin = () => {
    if (gameState === "playing") {
      setGameState("finished");
      sendMessage("game_result", { result: "win", winnings: 20 });
      setGameNotifications((prev) => [
        ...prev,
        "Congratulations! You won 20 diamonds!",
      ]);
    }
  };

  const simulateLoss = () => {
    if (gameState === "playing") {
      setGameState("finished");
      sendMessage("game_result", { result: "loss", winnings: 0 });
      setGameNotifications((prev) => [...prev, "Better luck next time!"]);
    }
  };

  const resetGame = () => {
    setGameState("idle");
    setGameNotifications([]);
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-indigo-900 via-purple-900 to-pink-900 flex justify-center items-center">
        <div className="text-center">
          <div className="relative">
            <div className="animate-spin rounded-full h-16 w-16 border-t-2 border-b-2 border-cyan-400 mx-auto mb-4"></div>
            <div className="absolute inset-0 animate-ping rounded-full h-16 w-16 border border-cyan-400 opacity-20"></div>
          </div>
          <p className="text-white text-lg font-light">
            Entering the Cosmos...
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-blue-900 to-indigo-900 relative">
      {/* Animated Stars Background */}
      <div className="fixed inset-0 pointer-events-none">
        <div className="stars absolute inset-0"></div>
        <div className="twinkling absolute inset-0"></div>
        <div className="clouds absolute inset-0"></div>
      </div>

      {/* Header */}
      <div className="relative z-10 bg-black bg-opacity-20 backdrop-blur-md border-b border-cyan-500 border-opacity-30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex justify-between items-center py-4">
            <div className="flex items-center space-x-6">
              <h1 className="text-3xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-cyan-400 via-blue-300 to-purple-400 flex items-center space-x-3">
                <FaStar className="text-cyan-400 animate-pulse" />
                <span>CASLETTE</span>
              </h1>
              <div className="flex items-center space-x-2">
                <div
                  className={`w-3 h-3 rounded-full ${
                    isConnected ? "bg-green-400" : "bg-red-400"
                  } animate-pulse`}
                />
                <span className="text-sm text-cyan-100 font-light">
                  {isConnected ? "Starlink Active" : "Connection Lost"}
                </span>
              </div>
            </div>
            <div className="flex items-center space-x-4 flex-wrap">
              <div className="flex items-center space-x-3 bg-gradient-to-r from-yellow-500 to-orange-500 bg-opacity-20 backdrop-blur-sm rounded-xl px-4 py-2 border border-yellow-400 border-opacity-30">
                <FaGem className="text-yellow-300 text-xl animate-pulse" />
                <span className="text-white font-bold text-lg">
                  {currentBalance.toLocaleString()}
                </span>
              </div>
              <span className="text-cyan-100 font-light hidden sm:block">
                Captain {user?.username}
              </span>
              <button
                onClick={logout}
                className="bg-gradient-to-r from-red-500 to-pink-500 hover:from-red-600 hover:to-pink-600 text-white px-4 py-2 rounded-xl font-medium transition-all duration-300 transform hover:scale-105 hover:shadow-lg"
              >
                Disconnect
              </button>
            </div>
          </div>
        </div>
      </div>

      {/* Main Game Area */}
      <div className="relative z-10 max-w-7xl mx-auto py-12 px-4 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
          {/* Central Game Console */}
          <div className="lg:col-span-2">
            <div className="bg-black bg-opacity-40 backdrop-blur-xl rounded-3xl border border-cyan-500 border-opacity-30 p-10 relative overflow-hidden">
              {/* Cosmic glow effect */}
              <div className="absolute inset-0 bg-gradient-to-r from-cyan-500 via-purple-500 to-pink-500 opacity-5 rounded-3xl"></div>

              <div className="relative z-10">
                <h2 className="text-4xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-cyan-300 via-blue-300 to-purple-300 mb-8 text-center flex items-center justify-center space-x-3">
                  <FaDice className="text-cyan-400 animate-bounce" />
                  <span>Stellar Fortune</span>
                </h2>

                <div className="text-center mb-12">
                  {/* Game State Orb */}
                  <div className="relative inline-block mb-8">
                    <div className="w-32 h-32 bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600 rounded-full flex items-center justify-center relative overflow-hidden glow-cyan">
                      <div className="absolute inset-0 bg-gradient-to-br from-cyan-400 via-blue-500 to-purple-600 rounded-full cosmic-pulse"></div>
                      <div className="relative z-10 text-6xl text-white float">
                        {gameState === "idle" && (
                          <FaBullseye className="animate-pulse" />
                        )}
                        {gameState === "playing" && (
                          <FaDice className="animate-spin" />
                        )}
                        {gameState === "finished" && (
                          <FaTrophy className="animate-bounce" />
                        )}
                      </div>
                    </div>
                    {/* Enhanced orbital rings */}
                    <div className="absolute inset-0 w-32 h-32 border-2 border-cyan-400 border-opacity-40 rounded-full animate-spin glow-cyan"></div>
                    <div
                      className="absolute -inset-2 w-36 h-36 border border-purple-400 border-opacity-30 rounded-full animate-spin glow-purple"
                      style={{
                        animationDuration: "3s",
                        animationDirection: "reverse",
                      }}
                    ></div>
                    <div
                      className="absolute -inset-4 w-40 h-40 border border-pink-400 border-opacity-20 rounded-full animate-spin"
                      style={{ animationDuration: "4s" }}
                    ></div>
                  </div>{" "}
                  <div className="text-white space-y-6">
                    <p className="text-xl font-light text-cyan-100">
                      {gameState === "idle" &&
                        "The stars align... Ready to venture into fortune?"}
                      {gameState === "playing" &&
                        "The cosmic dice are rolling through space..."}
                      {gameState === "finished" &&
                        "Your destiny has been written in the stars!"}
                    </p>

                    <div className="space-y-4">
                      {gameState === "idle" && (
                        <button
                          onClick={startGame}
                          disabled={currentBalance < 10 || !isConnected}
                          className="group relative bg-gradient-to-r from-cyan-500 to-blue-600 hover:from-cyan-600 hover:to-blue-700 disabled:from-gray-600 disabled:to-gray-700 disabled:cursor-not-allowed text-white px-12 py-4 rounded-xl font-bold text-xl transition-all duration-300 transform hover:scale-105 hover:shadow-xl hover:shadow-cyan-500/25 disabled:transform-none glow-cyan shimmer overflow-hidden"
                        >
                          <div className="absolute inset-0 bg-gradient-to-r from-cyan-400 to-blue-500 opacity-0 group-hover:opacity-20 transition-opacity duration-300"></div>
                          <span className="relative flex items-center space-x-3">
                            <FaGem className="cosmic-pulse" />
                            <span>Launch Mission (10 Gems)</span>
                            <FaStar className="float" />
                          </span>
                        </button>
                      )}

                      {gameState === "playing" && (
                        <div className="space-x-6">
                          <button
                            onClick={simulateWin}
                            className="group relative bg-gradient-to-r from-green-500 to-emerald-600 hover:from-green-600 hover:to-emerald-700 text-white px-8 py-3 rounded-xl font-semibold transition-all duration-300 transform hover:scale-105 hover:shadow-lg glow-cyan overflow-hidden"
                          >
                            <div className="absolute inset-0 bg-gradient-to-r from-green-400 to-emerald-500 opacity-0 group-hover:opacity-20 transition-opacity duration-300"></div>
                            <span className="relative flex items-center space-x-2">
                              <FaTrophy className="animate-pulse" />
                              <span>Victory Path</span>
                            </span>
                          </button>
                          <button
                            onClick={simulateLoss}
                            className="group relative bg-gradient-to-r from-red-500 to-pink-600 hover:from-red-600 hover:to-pink-700 text-white px-8 py-3 rounded-xl font-semibold transition-all duration-300 transform hover:scale-105 hover:shadow-lg overflow-hidden"
                          >
                            <div className="absolute inset-0 bg-gradient-to-r from-red-400 to-pink-500 opacity-0 group-hover:opacity-20 transition-opacity duration-300"></div>
                            <span className="relative flex items-center space-x-2">
                              <FaDice className="animate-spin" />
                              <span>Cosmic Trial</span>
                            </span>
                          </button>
                        </div>
                      )}

                      {gameState === "finished" && (
                        <button
                          onClick={resetGame}
                          className="group relative bg-gradient-to-r from-purple-500 to-indigo-600 hover:from-purple-600 hover:to-indigo-700 text-white px-12 py-4 rounded-xl font-bold text-xl transition-all duration-300 transform hover:scale-105 hover:shadow-xl hover:shadow-purple-500/25 glow-purple overflow-hidden"
                        >
                          <div className="absolute inset-0 bg-gradient-to-r from-purple-400 to-indigo-500 opacity-0 group-hover:opacity-20 transition-opacity duration-300"></div>
                          <span className="relative flex items-center space-x-3">
                            <FaStar className="cosmic-pulse" />
                            <span>New Constellation</span>
                            <FaBullseye className="float" />
                          </span>
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Side Panels */}
          <div className="space-y-8">
            {/* Cosmic Activity Feed */}
            <div className="bg-black bg-opacity-40 backdrop-blur-xl rounded-2xl border border-cyan-500 border-opacity-30 p-6 relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-cyan-500 via-purple-500 to-pink-500 opacity-5 rounded-2xl"></div>
              <div className="relative z-10">
                <h3 className="text-xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-cyan-300 to-purple-300 mb-4 flex items-center space-x-2">
                  <FaStar className="text-cyan-400" />
                  <span>Stellar Communications</span>
                </h3>
                <div className="space-y-3 max-h-64 overflow-y-auto custom-scrollbar">
                  {gameNotifications.length === 0 ? (
                    <p className="text-cyan-100 opacity-60 text-sm font-light">
                      Awaiting cosmic signals...
                    </p>
                  ) : (
                    gameNotifications
                      .slice(-10)
                      .reverse()
                      .map((notification, index) => (
                        <div
                          key={index}
                          className="bg-gradient-to-r from-cyan-500 to-purple-500 bg-opacity-10 backdrop-blur-sm rounded-xl p-3 border border-cyan-500 border-opacity-20"
                        >
                          <p className="text-cyan-100 text-sm font-light">
                            {notification}
                          </p>
                        </div>
                      ))
                  )}
                </div>
              </div>
            </div>

            {/* Transaction History */}
            <div className="bg-black bg-opacity-40 backdrop-blur-xl rounded-2xl border border-cyan-500 border-opacity-30 p-6 relative overflow-hidden">
              <div className="absolute inset-0 bg-gradient-to-br from-yellow-500 via-orange-500 to-red-500 opacity-5 rounded-2xl"></div>
              <div className="relative z-10">
                <h3 className="text-xl font-bold text-transparent bg-clip-text bg-gradient-to-r from-yellow-300 to-orange-300 mb-4 flex items-center space-x-2">
                  <FaGem className="text-yellow-400" />
                  <span>Gem Transactions</span>
                </h3>
                <div className="space-y-3 max-h-48 overflow-y-auto custom-scrollbar">
                  {diamonds
                    .slice(-5)
                    .reverse()
                    .map((diamond) => (
                      <div
                        key={diamond.id}
                        className="bg-gradient-to-r from-yellow-500 to-orange-500 bg-opacity-10 backdrop-blur-sm rounded-xl p-3 border border-yellow-500 border-opacity-20 flex justify-between items-center"
                      >
                        <div>
                          <p className="text-yellow-100 text-sm font-medium capitalize">
                            {diamond.type}
                          </p>
                          <p className="text-yellow-100 opacity-60 text-xs font-light">
                            {diamond.description}
                          </p>
                        </div>
                        <span
                          className={`font-bold text-lg ${
                            diamond.type === "credit"
                              ? "text-green-400"
                              : "text-red-400"
                          }`}
                        >
                          {diamond.type === "credit" ? "+" : "-"}
                          {Math.abs(diamond.amount)}
                        </span>
                      </div>
                    ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
};

export default HomePage;
