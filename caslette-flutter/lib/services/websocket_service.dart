import "package:web_socket_channel/web_socket_channel.dart";
import "package:flutter_riverpod/flutter_riverpod.dart";
import "dart:convert";

class WebSocketService {
  WebSocketChannel? _channel;
  String? _token;

  void connect(String token) {
    _token = token;
    try {
      // Use the same WebSocket URL as React app with token query parameter
      final wsUrl = "ws://localhost:8080/api/v1/ws?token=$token";
      _channel = WebSocketChannel.connect(Uri.parse(wsUrl));
      print("WebSocket connected to: ws://localhost:8080/api/v1/ws");
    } catch (e) {
      print("WebSocket connection failed: " + e.toString());
    }
  }

  void disconnect() {
    _channel?.sink.close();
    _channel = null;
    _token = null;
  }

  void send(Map<String, dynamic> message) {
    if (_channel != null) {
      final messageWithTimestamp = {
        ...message,
        'timestamp': DateTime.now().toIso8601String(),
      };
      _channel!.sink.add(jsonEncode(messageWithTimestamp));
    }
  }

  Stream<dynamic>? get stream => _channel?.stream;

  bool get isConnected => _channel != null;
}

final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  return WebSocketService();
});
