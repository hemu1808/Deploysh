type ConnectionCallback = () => void;
type MessageCallback = (data: any) => void;

const WS_BASE_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws';

class WebSocketService {
  private socket: WebSocket | null = null;
  private listeners: Set<MessageCallback> = new Set();
  private reconnectTimer: NodeJS.Timeout | null = null;
  private onConnectCb: ConnectionCallback | null = null;

  connect(onConnect: ConnectionCallback) {
    this.onConnectCb = onConnect;
    this.createSocket();
  }
  
  private createSocket() {
      this.socket = new WebSocket(WS_BASE_URL);

      this.socket.onopen = () => {
          console.log('[WebSocket] Connected to Go Backend');
          if (this.onConnectCb) this.onConnectCb();
          if (this.reconnectTimer) {
              clearTimeout(this.reconnectTimer);
              this.reconnectTimer = null;
          }
      };

      this.socket.onmessage = (event) => {
          try {
              const msg = JSON.parse(event.data);
              this.listeners.forEach(cb => cb(msg));
          } catch (e) {
              console.error("[WebSocket] Message parsing error:", e);
          }
      };

      this.socket.onclose = (event) => {
          console.log('[WebSocket] Disconnected from Go Backend', 'code', event.code, 'reason', event.reason);
          this.reconnectTimer = setTimeout(() => this.createSocket(), 3000);
      };

      this.socket.onerror = (err) => {
          console.error('[WebSocket] Connection Error:', err);
      }
  }

  disconnect() {
    if (this.socket) {
        // Prevent reconnect loop on intentional disconnect
        this.socket.onclose = null; 
        this.socket.close();
        this.socket = null;
    }
    if (this.reconnectTimer) {
        clearTimeout(this.reconnectTimer);
        this.reconnectTimer = null;
    }
  }

  subscribe(callback: MessageCallback) {
    this.listeners.add(callback);
    return () => this.listeners.delete(callback);
  }
}

export const wsService = new WebSocketService();
