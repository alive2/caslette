export interface WebSocketMessage {
  type: string;
  payload: any;
  timestamp: string;
}

export interface GameMessage extends WebSocketMessage {
  type: "game_update" | "balance_update" | "notification";
  payload: {
    balance?: number;
    game_state?: any;
    message?: string;
    diamonds?: any;
  };
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private token: string | null = null;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectInterval = 3000;
  private listeners: { [key: string]: ((data: any) => void)[] } = {};

  constructor(url = "ws://localhost:8080/api/v1/ws") {
    this.url = url;
  }

  connect(token: string): Promise<void> {
    return new Promise((resolve, reject) => {
      this.token = token;

      // Add token as query parameter for authentication
      const wsUrl = `${this.url}?token=${token}`;
      console.log(
        "🔌 Connecting to WebSocket:",
        wsUrl.replace(token, token.substring(0, 20) + "...")
      );

      try {
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          console.log("🎉 WebSocket connection opened successfully");
          this.reconnectAttempts = 0;
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const message: WebSocketMessage = JSON.parse(event.data);
            console.log("📨 Received WebSocket message:", message);
            this.handleMessage(message);
          } catch (error) {
            console.error("❌ Failed to parse WebSocket message:", error);
          }
        };

        this.ws.onclose = (event) => {
          console.log(
            "🔌 WebSocket disconnected. Code:",
            event.code,
            "Reason:",
            event.reason
          );
          this.handleReconnect();
        };

        this.ws.onerror = (error) => {
          console.error("❌ WebSocket error occurred:", error);
          reject(error);
        };
      } catch (error) {
        console.error("❌ Failed to create WebSocket connection:", error);
        reject(error);
      }
    });
  }

  disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.reconnectAttempts = this.maxReconnectAttempts; // Prevent reconnection
  }

  send(type: string, payload: any) {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      const message: WebSocketMessage = {
        type,
        payload,
        timestamp: new Date().toISOString(),
      };
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn("WebSocket not connected");
    }
  }

  on(eventType: string, callback: (data: any) => void) {
    if (!this.listeners[eventType]) {
      this.listeners[eventType] = [];
    }
    this.listeners[eventType].push(callback);
  }

  off(eventType: string, callback: (data: any) => void) {
    if (this.listeners[eventType]) {
      this.listeners[eventType] = this.listeners[eventType].filter(
        (listener) => listener !== callback
      );
    }
  }

  private handleMessage(message: WebSocketMessage) {
    // Emit to specific type listeners
    if (this.listeners[message.type]) {
      this.listeners[message.type].forEach((callback) => {
        callback(message.payload);
      });
    }

    // Emit to general message listeners
    if (this.listeners["message"]) {
      this.listeners["message"].forEach((callback) => {
        callback(message);
      });
    }
  }

  private handleReconnect() {
    if (this.reconnectAttempts < this.maxReconnectAttempts && this.token) {
      this.reconnectAttempts++;
      console.log(
        `Attempting to reconnect... (${this.reconnectAttempts}/${this.maxReconnectAttempts})`
      );

      setTimeout(() => {
        this.connect(this.token!).catch((error) => {
          console.error("Reconnection failed:", error);
        });
      }, this.reconnectInterval);
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

// Create a singleton instance
export const wsService = new WebSocketService();
