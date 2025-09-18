import 'dart:async';
import 'dart:convert';
import 'dart:developer';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:uuid/uuid.dart';

// WebSocket connection states
enum WebSocketConnectionState {
  disconnected,
  connecting,
  connected,
  reconnecting,
  error,
}

// WebSocket connection info
class WebSocketConnectionInfo {
  final WebSocketConnectionState state;
  final String? error;
  final bool isAuthenticated;
  final String? userID;
  final String? username;

  const WebSocketConnectionInfo({
    required this.state,
    this.error,
    this.isAuthenticated = false,
    this.userID,
    this.username,
  });

  WebSocketConnectionInfo copyWith({
    WebSocketConnectionState? state,
    String? error,
    bool? isAuthenticated,
    String? userID,
    String? username,
  }) {
    return WebSocketConnectionInfo(
      state: state ?? this.state,
      error: error ?? this.error,
      isAuthenticated: isAuthenticated ?? this.isAuthenticated,
      userID: userID ?? this.userID,
      username: username ?? this.username,
    );
  }
}

// WebSocket message
class WebSocketMessage {
  final String type;
  final String? event;
  final dynamic data;
  final String? room;
  final String? requestId;
  final bool? success;
  final String? error;
  final int timestamp;

  WebSocketMessage({
    required this.type,
    this.event,
    this.data,
    this.room,
    this.requestId,
    this.success,
    this.error,
    required this.timestamp,
  });

  Map<String, dynamic> toJson() {
    final Map<String, dynamic> json = {'type': type, 'timestamp': timestamp};

    if (event != null) json['event'] = event;
    if (data != null) json['data'] = data;
    if (room != null) json['room'] = room;
    if (requestId != null) json['requestId'] = requestId;
    if (success != null) json['success'] = success;
    if (error != null) json['error'] = error;

    return json;
  }

  factory WebSocketMessage.fromJson(Map<String, dynamic> json) {
    return WebSocketMessage(
      type: json['type'] ?? '',
      event: json['event'],
      data: json['data'],
      room: json['room'],
      requestId: json['requestId'],
      success: json['success'],
      error: json['error'],
      timestamp: json['timestamp'] ?? DateTime.now().millisecondsSinceEpoch,
    );
  }
}

// WebSocket event
class WebSocketEvent {
  final String type;
  final String? event;
  final dynamic data;
  final String? room;

  const WebSocketEvent({required this.type, this.event, this.data, this.room});
}

// WebSocket response for request-response pattern
class WebSocketResponse {
  final String type;
  final bool success;
  final dynamic data;
  final String? error;

  const WebSocketResponse({
    required this.type,
    required this.success,
    this.data,
    this.error,
  });

  factory WebSocketResponse.fromMessage(WebSocketMessage message) {
    return WebSocketResponse(
      type: message.type,
      success: message.success ?? false,
      data: message.data,
      error: message.error,
    );
  }
}

// Main WebSocket service
class WebSocketService {
  static const String _baseUrl = 'ws://localhost:8081/ws';

  WebSocketChannel? _channel;
  final Map<String, Completer<WebSocketResponse>> _pendingRequests = {};
  final Map<String, Function(WebSocketEvent)> _eventHandlers = {};

  // Connection state stream
  final StreamController<WebSocketConnectionInfo> _connectionController =
      StreamController<WebSocketConnectionInfo>.broadcast();

  // Event stream for server-initiated events
  final StreamController<WebSocketEvent> _eventController =
      StreamController<WebSocketEvent>.broadcast();

  // Reconnection management
  Timer? _reconnectionTimer;
  int _reconnectionAttempts = 0;
  static const int _maxReconnectionAttempts = 10;
  static const Duration _initialReconnectionDelay = Duration(seconds: 1);

  WebSocketConnectionInfo _connectionInfo = const WebSocketConnectionInfo(
    state: WebSocketConnectionState.disconnected,
  );

  final Uuid _uuid = const Uuid();

  // Constructor
  WebSocketService() {
    // Emit initial state immediately so UI doesn't show "Loading..."
    Future.microtask(() {
      _connectionController.add(_connectionInfo);
    });
  }

  // Getters
  Stream<WebSocketConnectionInfo> get connectionStream =>
      _connectionController.stream;
  Stream<WebSocketEvent> get eventStream => _eventController.stream;
  WebSocketConnectionInfo get connectionInfo => _connectionInfo;
  bool get isConnected =>
      _connectionInfo.state == WebSocketConnectionState.connected;
  bool get isAuthenticated => _connectionInfo.isAuthenticated;

  // Connect to WebSocket server
  Future<void> connect({String? customUrl}) async {
    if (isConnected) {
      log('WebSocket already connected');
      return;
    }

    try {
      _updateConnectionState(WebSocketConnectionState.connecting);
      _reconnectionAttempts = 0;

      final uri = Uri.parse(customUrl ?? _baseUrl);
      _channel = WebSocketChannel.connect(uri);

      // Listen to messages
      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDisconnection,
      );

      // Wait for connection to be established
      await _waitForConnection(timeout: const Duration(seconds: 10));
    } catch (e) {
      log('WebSocket connection error: $e');
      _updateConnectionState(
        WebSocketConnectionState.error,
        error: e.toString(),
      );
      _scheduleReconnection();
      rethrow;
    }
  }

  // Disconnect from WebSocket server
  Future<void> disconnect() async {
    _reconnectionTimer?.cancel();
    _reconnectionTimer = null;

    if (_channel != null) {
      await _channel!.sink.close();
      _channel = null;
    }

    _updateConnectionState(WebSocketConnectionState.disconnected);

    // Clear pending requests
    for (final completer in _pendingRequests.values) {
      if (!completer.isCompleted) {
        completer.completeError('Connection closed');
      }
    }
    _pendingRequests.clear();
  }

  // Authenticate with JWT token
  Future<WebSocketResponse> authenticate(String token) async {
    log('WebSocket authenticate called with token length: ${token.length}');
    log('Current connection state: $isConnected');

    final response = await sendRequest('auth', {'token': token});
    log(
      'Authentication response received: success=${response.success}, error=${response.error}',
    );

    if (response.success) {
      final data = response.data as Map<String, dynamic>?;
      log(
        'Authentication successful, updating connection state with user data: $data',
      );
      _updateConnectionState(
        WebSocketConnectionState.connected,
        isAuthenticated: true,
        userID: data?['userID'],
        username: data?['username'],
      );
    } else {
      log('Authentication failed: ${response.error}');
    }

    return response;
  }

  // Join a room
  Future<WebSocketResponse> joinRoom(String room) async {
    print('WebSocketService: joinRoom called for room: $room');
    print('WebSocketService: About to send join_room request...');
    final response = await sendRequest('join_room', room);
    print(
      'WebSocketService: joinRoom response received - type: ${response.type}, success: ${response.success}, error: ${response.error}, data: ${response.data}',
    );
    return response;
  }

  // Leave a room
  Future<WebSocketResponse> leaveRoom(String room) async {
    return await sendRequest('leave_room', room);
  }

  // Send message to a room
  Future<WebSocketResponse> sendToRoom(String room, dynamic message) async {
    return await sendRequest('send_to_room', {
      'room': room,
      'message': message,
    });
  }

  // Send a request and wait for response
  Future<WebSocketResponse> sendRequest(String type, dynamic data) async {
    if (!isConnected) {
      throw Exception('WebSocket not connected');
    }

    final requestId = _uuid.v4();
    final message = WebSocketMessage(
      type: type,
      data: data,
      requestId: requestId,
      timestamp: DateTime.now().millisecondsSinceEpoch,
    );

    final completer = Completer<WebSocketResponse>();
    _pendingRequests[requestId] = completer;

    // Send the message
    _sendMessage(message);

    // Set timeout for the request
    Timer(const Duration(seconds: 30), () {
      if (_pendingRequests.containsKey(requestId)) {
        _pendingRequests.remove(requestId);
        if (!completer.isCompleted) {
          completer.completeError('Request timeout');
        }
      }
    });

    return completer.future;
  }

  // Send a message without expecting a response
  void sendMessage(String type, dynamic data, {String? room}) {
    if (!isConnected) {
      log('Cannot send message: WebSocket not connected');
      return;
    }

    final message = WebSocketMessage(
      type: type,
      data: data,
      room: room,
      timestamp: DateTime.now().millisecondsSinceEpoch,
    );

    _sendMessage(message);
  }

  // Register an event handler
  void on(String eventType, Function(WebSocketEvent) handler) {
    _eventHandlers[eventType] = handler;
  }

  // Remove an event handler
  void off(String eventType) {
    _eventHandlers.remove(eventType);
  }

  // Manual reconnection
  Future<void> reconnect() async {
    await disconnect();
    await connect();
  }

  // Dispose resources
  void dispose() {
    disconnect();
    _connectionController.close();
    _eventController.close();
  }

  // Private methods
  void _sendMessage(WebSocketMessage message) {
    if (_channel != null) {
      final jsonString = jsonEncode(message.toJson());
      _channel!.sink.add(jsonString);
      log('WebSocket sent: $jsonString');
    }
  }

  void _handleMessage(dynamic messageData) {
    try {
      final Map<String, dynamic> json = jsonDecode(messageData);
      final message = WebSocketMessage.fromJson(json);

      print('WebSocket received: ${message.type} - $messageData');

      // Handle responses to requests
      if (message.requestId != null &&
          _pendingRequests.containsKey(message.requestId)) {
        print(
          'WebSocket: Processing response for request ${message.requestId}',
        );
        print(
          'WebSocket: Response type=${message.type}, success=${message.success}, error=${message.error}',
        );

        final completer = _pendingRequests.remove(message.requestId!);
        if (completer != null && !completer.isCompleted) {
          final response = WebSocketResponse.fromMessage(message);
          print(
            'WebSocket: Completing request with response - success: ${response.success}, error: ${response.error}',
          );
          completer.complete(response);
        }
        return;
      }

      // Handle specific message types
      switch (message.type) {
        case 'connected':
          _updateConnectionState(WebSocketConnectionState.connected);
          break;
        case 'error':
          log('WebSocket error message: ${message.error}');
          break;
        default:
          // Handle as event
          final event = WebSocketEvent(
            type: message.type,
            event: message.event,
            data: message.data,
            room: message.room,
          );

          _eventController.add(event);

          // Call registered handlers
          final handler = _eventHandlers[message.type];
          if (handler != null) {
            handler(event);
          }
          break;
      }
    } catch (e) {
      log('Error handling WebSocket message: $e');
    }
  }

  void _handleError(error) {
    log('WebSocket error: $error');
    _updateConnectionState(
      WebSocketConnectionState.error,
      error: error.toString(),
    );
    _scheduleReconnection();
  }

  void _handleDisconnection() {
    log('WebSocket disconnected');
    _updateConnectionState(WebSocketConnectionState.disconnected);
    _scheduleReconnection();
  }

  void _updateConnectionState(
    WebSocketConnectionState state, {
    String? error,
    bool? isAuthenticated,
    String? userID,
    String? username,
  }) {
    _connectionInfo = _connectionInfo.copyWith(
      state: state,
      error: error,
      isAuthenticated: isAuthenticated,
      userID: userID,
      username: username,
    );
    _connectionController.add(_connectionInfo);
  }

  Future<void> _waitForConnection({required Duration timeout}) async {
    final completer = Completer<void>();
    late StreamSubscription<WebSocketConnectionInfo> subscription;

    subscription = connectionStream.listen((info) {
      if (info.state == WebSocketConnectionState.connected) {
        if (!completer.isCompleted) {
          completer.complete();
        }
        subscription.cancel();
      } else if (info.state == WebSocketConnectionState.error) {
        if (!completer.isCompleted) {
          completer.completeError(info.error ?? 'Connection failed');
        }
        subscription.cancel();
      }
    });

    // Set timeout
    Timer(timeout, () {
      if (!completer.isCompleted) {
        subscription.cancel();
        completer.completeError('Connection timeout');
      }
    });

    return completer.future;
  }

  void _scheduleReconnection() {
    if (_reconnectionAttempts >= _maxReconnectionAttempts) {
      log('Max reconnection attempts reached');
      return;
    }

    _reconnectionTimer?.cancel();

    final delay = Duration(
      milliseconds:
          (_initialReconnectionDelay.inMilliseconds *
                  (1 << _reconnectionAttempts))
              .clamp(1000, 30000),
    );

    log(
      'Scheduling reconnection in ${delay.inMilliseconds}ms (attempt ${_reconnectionAttempts + 1})',
    );

    _updateConnectionState(WebSocketConnectionState.reconnecting);

    _reconnectionTimer = Timer(delay, () {
      _reconnectionAttempts++;
      connect().catchError((e) {
        log('Reconnection failed: $e');
      });
    });
  }
}
