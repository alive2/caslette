import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'dart:developer';
import 'websocket_service.dart';

// WebSocket service provider
final webSocketServiceProvider = Provider<WebSocketService>((ref) {
  return WebSocketService();
});

// WebSocket connection state provider
final webSocketConnectionProvider = StreamProvider<WebSocketConnectionInfo>((
  ref,
) {
  final service = ref.watch(webSocketServiceProvider);
  return service.connectionStream;
});

// WebSocket events provider
final webSocketEventsProvider = StreamProvider<WebSocketEvent>((ref) {
  final service = ref.watch(webSocketServiceProvider);
  return service.eventStream;
});

// Authentication state provider
final isAuthenticatedProvider = Provider<bool>((ref) {
  final connectionAsync = ref.watch(webSocketConnectionProvider);
  return connectionAsync.when(
    data: (info) => info.isAuthenticated,
    loading: () => false,
    error: (_, __) => false,
  );
});

// WebSocket controller for managing connection and authentication
class WebSocketController extends StateNotifier<AsyncValue<void>> {
  final WebSocketService _service;

  WebSocketController(this._service) : super(const AsyncValue.data(null));

  // Connect to WebSocket
  Future<void> connect() async {
    state = const AsyncValue.loading();
    try {
      await _service.connect();
      state = const AsyncValue.data(null);
    } catch (e, stack) {
      log('WebSocket connection failed: $e');
      state = AsyncValue.error(e, stack);
    }
  }

  // Disconnect from WebSocket
  Future<void> disconnect() async {
    try {
      await _service.disconnect();
      state = const AsyncValue.data(null);
    } catch (e, stack) {
      log('WebSocket disconnection failed: $e');
      state = AsyncValue.error(e, stack);
    }
  }

  // Authenticate with stored token
  Future<bool> authenticateWithStoredToken() async {
    try {
      log('WebSocketController: authenticateWithStoredToken called');
      final prefs = await SharedPreferences.getInstance();
      final token = prefs.getString('auth_token');

      if (token == null) {
        log('WebSocketController: No stored token found');
        return false;
      }

      log(
        'WebSocketController: Found stored token, attempting authentication...',
      );
      final response = await _service.authenticate(token);
      if (response.success) {
        log('WebSocketController: Authentication successful');
        return true;
      } else {
        log('WebSocketController: Authentication failed: ${response.error}');
        // Clear invalid token
        await prefs.remove('auth_token');
        return false;
      }
    } catch (e) {
      log('WebSocketController: Authentication error: $e');
      return false;
    }
  }

  // Authenticate with provided token
  Future<bool> authenticate(String token) async {
    try {
      log(
        'WebSocketController: authenticate called with token length: ${token.length}',
      );
      final response = await _service.authenticate(token);
      if (response.success) {
        // Store the token
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('auth_token', token);
        log('WebSocketController: Authentication successful, token stored');
        return true;
      } else {
        log('WebSocketController: Authentication failed: ${response.error}');
        return false;
      }
    } catch (e) {
      log('WebSocketController: Authentication error: $e');
      return false;
    }
  }

  // Reconnect with authentication
  Future<void> reconnectWithAuth() async {
    await connect();
    await authenticateWithStoredToken();
  }

  // Get user balance via WebSocket
  Future<Map<String, dynamic>?> getUserBalance(String userId) async {
    try {
      log('WebSocketController: getUserBalance called for userId: $userId');
      final response = await _service.getUserBalance(userId);
      if (response.success) {
        log('WebSocketController: getUserBalance successful');
        return response.data as Map<String, dynamic>?;
      } else {
        log('WebSocketController: getUserBalance failed: ${response.error}');
        return null;
      }
    } catch (e) {
      log('WebSocketController: getUserBalance error: $e');
      return null;
    }
  }

  // Get user profile via WebSocket
  Future<Map<String, dynamic>?> getUserProfile(String userId) async {
    try {
      log('WebSocketController: getUserProfile called for userId: $userId');
      final response = await _service.getUserProfile(userId);
      if (response.success) {
        log('WebSocketController: getUserProfile successful');
        return response.data as Map<String, dynamic>?;
      } else {
        log('WebSocketController: getUserProfile failed: ${response.error}');
        return null;
      }
    } catch (e) {
      log('WebSocketController: getUserProfile error: $e');
      return null;
    }
  }
}

// WebSocket controller provider
final webSocketControllerProvider =
    StateNotifierProvider<WebSocketController, AsyncValue<void>>((ref) {
      final service = ref.watch(webSocketServiceProvider);
      return WebSocketController(service);
    });

// Room management controller
class RoomController extends StateNotifier<Set<String>> {
  final WebSocketService _service;

  RoomController(this._service) : super(<String>{});

  // Join a room
  Future<bool> joinRoom(String room) async {
    try {
      print('RoomController: Attempting to join room: $room');
      final response = await _service.joinRoom(room);
      print(
        'RoomController: Join room response received - type: ${response.type}, success: ${response.success}, error: ${response.error}, data: ${response.data}',
      );

      if (response.success) {
        state = {...state, room};
        print(
          'RoomController: Successfully joined room: $room, current rooms: $state',
        );
        return true;
      } else {
        print('RoomController: Failed to join room $room: ${response.error}');
        return false;
      }
    } catch (e) {
      print('RoomController: Exception joining room $room: $e');
      return false;
    }
  }

  // Handle auto-join when creating a room
  void handleAutoJoin(String room) {
    print('RoomController: Auto-joining room: $room');
    state = {...state, room};
    print('RoomController: Auto-joined room: $room, current rooms: $state');
  }

  // Leave a room
  Future<bool> leaveRoom(String room) async {
    try {
      final response = await _service.leaveRoom(room);
      if (response.success) {
        state = state.where((r) => r != room).toSet();
        log('Left room: $room');
        return true;
      } else {
        log('Failed to leave room $room: ${response.error}');
        return false;
      }
    } catch (e) {
      log('Error leaving room $room: $e');
      return false;
    }
  }

  // Send message to room
  Future<bool> sendToRoom(String room, dynamic message) async {
    try {
      print(
        'RoomController: Sending message to room: $room, message: $message',
      );
      final response = await _service.sendToRoom(room, message);
      print(
        'RoomController: Send to room response - success: ${response.success}, error: ${response.error}, data: ${response.data}',
      );
      if (response.success) {
        print('Message sent to room $room');
        return true;
      } else {
        print('Failed to send message to room $room: ${response.error}');
        return false;
      }
    } catch (e) {
      print('Error sending message to room $room: $e');
      return false;
    }
  }

  // Get joined rooms
  Set<String> get joinedRooms => state;
}

// Room controller provider
final roomControllerProvider =
    StateNotifierProvider<RoomController, Set<String>>((ref) {
      final service = ref.watch(webSocketServiceProvider);
      return RoomController(service);
    });

// Current user info provider (from WebSocket authentication)
final currentUserProvider = Provider<Map<String, dynamic>?>((ref) {
  final connectionAsync = ref.watch(webSocketConnectionProvider);
  return connectionAsync.when(
    data: (info) => info.isAuthenticated
        ? {'userID': info.userID, 'username': info.username}
        : null,
    loading: () => null,
    error: (_, __) => null,
  );
});

// Connection status provider for UI
final connectionStatusProvider = Provider<String>((ref) {
  final connectionAsync = ref.watch(webSocketConnectionProvider);
  return connectionAsync.when(
    data: (info) {
      switch (info.state) {
        case WebSocketConnectionState.disconnected:
          return 'Disconnected';
        case WebSocketConnectionState.connecting:
          return 'Connecting...';
        case WebSocketConnectionState.connected:
          return info.isAuthenticated
              ? 'Connected & Authenticated'
              : 'Connected';
        case WebSocketConnectionState.reconnecting:
          return 'Reconnecting...';
        case WebSocketConnectionState.error:
          return 'Error: ${info.error ?? 'Unknown error'}';
      }
    },
    loading: () => 'Loading...',
    error: (error, _) => 'Error: $error',
  );
});
