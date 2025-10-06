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
  final dynamic data;
  final String? room;
  final String? event;
  final String? requestId;
  final bool success;
  final String? error;
  final int timestamp;

  const WebSocketMessage({
    required this.type,
    this.data,
    this.room,
    this.event,
    this.requestId,
    this.success = false,
    this.error,
    this.timestamp = 0,
  });

  factory WebSocketMessage.fromJson(Map<String, dynamic> json) {
    return WebSocketMessage(
      type: json['type'] ?? '',
      data: json['data'],
      room: json['room'],
      event: json['event'],
      requestId: json['requestId'],
      success: json['success'] ?? false,
      error: json['error'],
      timestamp: json['timestamp'] ?? DateTime.now().millisecondsSinceEpoch,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type,
      'data': data,
      'room': room,
      'event': event,
      'requestId': requestId,
      'success': success,
      'error': error,
      'timestamp': timestamp,
    };
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
  final String? requestId;

  const WebSocketResponse({
    required this.type,
    required this.success,
    this.data,
    this.error,
    this.requestId,
  });

  factory WebSocketResponse.fromMessage(WebSocketMessage message) {
    return WebSocketResponse(
      type: message.type,
      success: message.success,
      data: message.data,
      error: message.error,
      requestId: message.requestId,
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
  bool _intentionalDisconnect = false; // Track if disconnect was intentional
  bool _allowReconnection = true; // Control whether reconnection is allowed

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
    if (!_allowReconnection) {
      log('WebSocket connection not allowed - user logged out');
      return;
    }

    if (isConnected) {
      log('WebSocket already connected');
      return;
    }

    try {
      // Cancel any pending reconnection attempts
      _reconnectionTimer?.cancel();
      _reconnectionTimer = null;

      // Reset intentional disconnect flag for new connections
      _intentionalDisconnect = false;

      _updateConnectionState(WebSocketConnectionState.connecting);
      _reconnectionAttempts = 0;

      final uri = Uri.parse(customUrl ?? _baseUrl);
      log('WebSocket connecting to: $uri');
      _channel = WebSocketChannel.connect(uri);

      // Start listening to messages immediately
      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDisconnection,
      );

      // Wait for connection to be established (server sends "connected" message)
      await _waitForConnection(timeout: const Duration(seconds: 10));

      log('WebSocket connected successfully');
    } catch (e) {
      log('WebSocket connection failed: $e');
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
    log('WebSocket manually disconnecting');

    // Mark this as an intentional disconnect and disable reconnection
    _intentionalDisconnect = true;
    _allowReconnection = false;

    // Send logout message to server if connected and authenticated
    if (isConnected && _connectionInfo.isAuthenticated) {
      try {
        log('Sending logout message to server');
        sendMessage('logout', {});
        // Give a brief moment for the message to be sent
        await Future.delayed(const Duration(milliseconds: 100));
      } catch (e) {
        log('Error sending logout message: $e');
      }
    }

    _reconnectionTimer?.cancel();
    _reconnectionTimer = null;

    if (_channel != null) {
      await _channel!.sink.close();
      _channel = null;
    }

    // Update connection state and clear authentication
    _updateConnectionState(
      WebSocketConnectionState.disconnected,
      isAuthenticated: false,
      userID: null,
      username: null,
    );

    // Clear pending requests
    for (final completer in _pendingRequests.values) {
      if (!completer.isCompleted) {
        completer.completeError('Connection closed');
      }
    }
    _pendingRequests.clear();
  }

  // Re-enable WebSocket connections (call when user logs in)
  void enableReconnection() {
    log('WebSocket reconnection re-enabled');
    _allowReconnection = true;
    _intentionalDisconnect = false;
  }

  // Authenticate with JWT token
  Future<WebSocketResponse> authenticate(String token) async {
    log('WebSocket authenticate called with token length: ${token.length}');
    print(
      'WebSocket AUTHENTICATE called with token: ${token.substring(0, 20)}...',
    );
    log('Current connection state: $isConnected');
    print('Current connection state: $isConnected');
    print('Channel is null: ${_channel == null}');

    // Check if channel exists (connection is established)
    if (_channel == null) {
      throw Exception('WebSocket channel not established');
    }

    print('WebSocket sending auth request...');
    final response = await sendRequest('auth', {'token': token});
    log(
      'Authentication response received: success=${response.success}, error=${response.error}',
    );
    print(
      'Authentication response: success=${response.success}, error=${response.error}',
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
    log('WebSocketService: joinRoom called for room: $room');
    final response = await sendRequest('join_room', room);
    log(
      'WebSocketService: joinRoom response - success: ${response.success}, error: ${response.error}',
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

  // Get user balance via WebSocket
  Future<WebSocketResponse> getUserBalance(String userId) async {
    log('WebSocketService: getUserBalance called for userId: $userId');
    final response = await sendRequest('get_user_balance', {'userId': userId});
    log(
      'WebSocketService: getUserBalance response - success: ${response.success}, error: ${response.error}',
    );
    return response;
  }

  // Get user profile via WebSocket
  Future<WebSocketResponse> getUserProfile(String userId) async {
    log('WebSocketService: getUserProfile called for userId: $userId');
    final response = await sendRequest('get_user_profile', {'userId': userId});
    log(
      'WebSocketService: getUserProfile response - success: ${response.success}, error: ${response.error}',
    );
    return response;
  }

  // Send request and wait for response
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

    print('Sending WebSocket request: $type with ID: $requestId');
    print('Request data: $data');

    final completer = Completer<WebSocketResponse>();
    _pendingRequests[requestId] = completer;

    // Send the message
    _sendMessage(message);

    // Set timeout for the request - reduced to 15 seconds for faster debugging
    Timer(const Duration(seconds: 15), () {
      if (_pendingRequests.containsKey(requestId)) {
        _pendingRequests.remove(requestId);
        if (!completer.isCompleted) {
          print('Request timeout for $type (ID: $requestId)');
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

  // Private methods
  void _sendMessage(WebSocketMessage message) {
    if (_channel == null) {
      log('Cannot send message: WebSocket channel is null');
      return;
    }

    final json = jsonEncode(message.toJson());
    log('WebSocket sending: ${message.type} - $json');
    print('WebSocket SENDING MESSAGE: ${message.type}');
    print('WebSocket JSON: $json');
    _channel!.sink.add(json);
  }

  void _handleMessage(dynamic messageData) {
    try {
      print('WebSocket RAW MESSAGE RECEIVED: $messageData');
      final Map<String, dynamic> json = jsonDecode(messageData);
      final message = WebSocketMessage.fromJson(json);

      log('WebSocket received: ${message.type} - $messageData');
      print(
        'WebSocket PARSED MESSAGE: type=${message.type}, requestId=${message.requestId}',
      );

      // Handle responses to requests
      if (message.requestId != null &&
          _pendingRequests.containsKey(message.requestId)) {
        print('WebSocket: Found pending request for ${message.requestId}');
        final completer = _pendingRequests.remove(message.requestId!);
        if (completer != null && !completer.isCompleted) {
          final response = WebSocketResponse.fromMessage(message);
          print(
            'WebSocket: Completing request ${message.requestId} with success=${response.success}',
          );
          completer.complete(response);
        }
        return;
      }

      // Handle specific message types
      switch (message.type) {
        case 'connected':
          print('WebSocket: Received connected message, updating state');
          _updateConnectionState(WebSocketConnectionState.connected);
          break;
        case 'error':
          log('WebSocket error message: ${message.error}');
          break;
        default:
          // Handle as event - emit to event stream
          final event = WebSocketEvent(
            type: message.type,
            event: message.event,
            data: message.data,
            room: message.room,
          );
          _eventController.add(event);

          // Call specific handler if registered
          final handler = _eventHandlers[message.type];
          if (handler != null) {
            handler(event);
          }
      }
    } catch (e) {
      log('Error handling WebSocket message: $e');
      print('WebSocket MESSAGE HANDLING ERROR: $e');
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
    log(
      'WebSocket disconnected (intentional: $_intentionalDisconnect, allowReconnect: $_allowReconnection)',
    );

    final wasAuthenticated = _connectionInfo.isAuthenticated;
    _updateConnectionState(WebSocketConnectionState.disconnected);

    // Only schedule reconnection if allowed and wasn't intentional
    if (_allowReconnection &&
        !_intentionalDisconnect &&
        wasAuthenticated &&
        _reconnectionTimer == null) {
      log('WebSocket scheduling reconnection - unintentional disconnect');
      _scheduleReconnection();
    } else {
      log('WebSocket NOT reconnecting');
      _intentionalDisconnect = false;
    }
  }

  void _updateConnectionState(
    WebSocketConnectionState state, {
    String? error,
    bool? isAuthenticated,
    String? userID,
    String? username,
  }) {
    log('WebSocket state change: ${_connectionInfo.state} -> $state');
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

    Timer(timeout, () {
      if (!completer.isCompleted) {
        subscription.cancel();
        completer.completeError('Connection timeout');
      }
    });

    return completer.future;
  }

  void _scheduleReconnection() {
    if (_reconnectionTimer != null ||
        _reconnectionAttempts >= _maxReconnectionAttempts) {
      return;
    }

    _reconnectionAttempts++;
    final delay = Duration(
      seconds: _initialReconnectionDelay.inSeconds * _reconnectionAttempts,
    );

    log(
      'WebSocket scheduling reconnection attempt $_reconnectionAttempts in ${delay.inSeconds}s',
    );
    _updateConnectionState(WebSocketConnectionState.reconnecting);

    _reconnectionTimer = Timer(delay, () {
      _reconnectionTimer = null;
      if (_connectionInfo.state != WebSocketConnectionState.connected) {
        log('WebSocket attempting reconnection $_reconnectionAttempts');
        connect().catchError((e) {
          log('WebSocket reconnection failed: $e');
          _scheduleReconnection();
        });
      }
    });
  }

  // Subscribe to events
  void on(String eventType, Function(WebSocketEvent) handler) {
    _eventHandlers[eventType] = handler;
  }

  // Unsubscribe from events
  void off(String eventType) {
    _eventHandlers.remove(eventType);
  }

  // Dispose resources
  void dispose() {
    _reconnectionTimer?.cancel();
    _channel?.sink.close();
    _connectionController.close();
    _eventController.close();
  }
}
