import React, { createContext, useContext, useEffect, useState } from "react";
import type { ReactNode } from "react";
import { wsService } from "../services/websocket";
import type { WebSocketMessage } from "../services/websocket";
import { useAuth } from "./AuthContext";

interface WebSocketContextType {
  isConnected: boolean;
  lastMessage: WebSocketMessage | null;
  sendMessage: (type: string, payload: Record<string, unknown>) => void;
  onMessage: (
    eventType: string,
    callback: (data: Record<string, unknown>) => void
  ) => void;
  offMessage: (
    eventType: string,
    callback: (data: Record<string, unknown>) => void
  ) => void;
}

const WebSocketContext = createContext<WebSocketContextType | undefined>(
  undefined
);

export const useWebSocket = () => {
  const context = useContext(WebSocketContext);
  if (context === undefined) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};

interface WebSocketProviderProps {
  children: ReactNode;
}

export const WebSocketProvider: React.FC<WebSocketProviderProps> = ({
  children,
}) => {
  const { token, user } = useAuth();
  const [isConnected, setIsConnected] = useState(false);
  const [lastMessage, setLastMessage] = useState<WebSocketMessage | null>(null);

  useEffect(() => {
    if (token && user) {
      console.log(
        "🚀 Attempting WebSocket connection for user:",
        user.username,
        "with token:",
        token?.substring(0, 20) + "..."
      );
      // Connect to WebSocket when user is authenticated
      wsService
        .connect(token)
        .then(() => {
          setIsConnected(true);
          console.log(
            "✅ WebSocket connected successfully for user:",
            user.username
          );
        })
        .catch((error) => {
          console.error("❌ Failed to connect to WebSocket:", error);
          console.error("Token used:", token?.substring(0, 20) + "...");
          setIsConnected(false);
        });

      // Listen for all messages
      const messageHandler = (message: WebSocketMessage) => {
        setLastMessage(message);
      };

      wsService.on("message", messageHandler);

      // Listen for connection status changes
      const connectionHandler = () => {
        setIsConnected(wsService.isConnected());
      };

      // Set up interval to check connection status
      const statusInterval = setInterval(connectionHandler, 1000);

      return () => {
        wsService.off("message", messageHandler);
        clearInterval(statusInterval);
        wsService.disconnect();
        setIsConnected(false);
      };
    } else {
      // Disconnect if user is not authenticated
      wsService.disconnect();
      setIsConnected(false);
    }
  }, [token, user]);

  const sendMessage = (type: string, payload: Record<string, unknown>) => {
    wsService.send(type, payload);
  };

  const onMessage = (
    eventType: string,
    callback: (data: Record<string, unknown>) => void
  ) => {
    wsService.on(eventType, callback);
  };

  const offMessage = (
    eventType: string,
    callback: (data: Record<string, unknown>) => void
  ) => {
    wsService.off(eventType, callback);
  };

  return (
    <WebSocketContext.Provider
      value={{
        isConnected,
        lastMessage,
        sendMessage,
        onMessage,
        offMessage,
      }}
    >
      {children}
    </WebSocketContext.Provider>
  );
};
