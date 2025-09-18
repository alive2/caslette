import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/websocket/websocket_providers.dart';
import 'dart:developer';
import 'dart:async';

class WebSocketTestPage extends ConsumerStatefulWidget {
  const WebSocketTestPage({super.key});

  @override
  ConsumerState<WebSocketTestPage> createState() => _WebSocketTestPageState();
}

class _WebSocketTestPageState extends ConsumerState<WebSocketTestPage> {
  final TextEditingController _tokenController = TextEditingController();
  final TextEditingController _roomController = TextEditingController();
  final TextEditingController _messageController = TextEditingController();
  final TextEditingController _requestTestController = TextEditingController();
  final List<String> _eventLog = [];
  final List<String> _requestResponses = [];
  final List<String> _subscriptionEvents = [];
  String? _lastRequestId;
  bool _isAwaitingResponse = false;

  @override
  void initState() {
    super.initState();
    // Listen to WebSocket events
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final service = ref.read(webSocketServiceProvider);
      service.eventStream.listen((event) {
        print('[WebSocket Test] ===== EVENT RECEIVED =====');
        print('[WebSocket Test] Event Type: ${event.type}');
        print('[WebSocket Test] Event Data: ${event.data}');
        print('[WebSocket Test] Event Room: ${event.room}');
        print('[WebSocket Test] ===============================');

        setState(() {
          _eventLog.add(
            '[${DateTime.now().toIso8601String()}] Event: ${event.type} - ${event.data}',
          );

          // Handle subscription events
          if (event.type == 'server_announcements' ||
              event.type == 'user_activity' ||
              event.type == 'broadcast') {
            _subscriptionEvents.add(
              '[${DateTime.now().toIso8601String()}] ${event.type}: ${event.data}',
            );
          }

          // Handle room response events
          if (event.type == 'create_room_response') {
            if (event.data != null && event.data['success'] == true) {
              _showSnackBar('Room created successfully: ${event.data['room']}');
            } else {
              _showSnackBar(
                'Failed to create room: ${event.data?['error'] ?? 'Unknown error'}',
              );
            }
          }

          if (event.type == 'list_rooms_response') {
            if (event.data != null && event.data['success'] == true) {
              final rooms = event.data['rooms'] as List?;
              final roomCount = rooms?.length ?? 0;
              _showSnackBar('Found $roomCount rooms');
              _logSubscriptionEvent(
                'Available rooms: ${rooms?.map((r) => r['name']).join(', ') ?? 'None'}',
              );
            } else {
              _showSnackBar(
                'Failed to list rooms: ${event.data?['error'] ?? 'Unknown error'}',
              );
            }
          }

          if (event.type == 'room_created') {
            _logSubscriptionEvent(
              'New room created: ${event.data?['room']} by ${event.data?['creator']}',
            );
          }
        });
        log('WebSocket event received: ${event.type}');
      });

      // Auto-connect and try authentication with stored token
      print('[WebSocket Test] Starting auto-connect and auth...');
      _autoConnectAndAuth();
    });
  }

  Future<void> _autoConnectAndAuth() async {
    try {
      print('[WebSocket Test] _autoConnectAndAuth started');

      // Load stored token if available
      final prefs = await SharedPreferences.getInstance();
      final storedToken = prefs.getString('auth_token');
      print(
        '[WebSocket Test] Stored token: ${storedToken != null ? 'Found (${storedToken.length} chars)' : 'None'}',
      );

      if (storedToken != null && storedToken.isNotEmpty) {
        _tokenController.text = storedToken;
        print('[WebSocket Test] Token loaded into text field');
        final preview = storedToken.length > 50
            ? storedToken.substring(0, 50) + '...'
            : storedToken;
        print('[WebSocket Test] Token preview: $preview');
      }

      // Auto-connect to WebSocket
      print('[WebSocket Test] Connecting to WebSocket...');
      await ref.read(webSocketControllerProvider.notifier).connect();
      print('[WebSocket Test] WebSocket connection attempt completed');

      // Wait a moment for connection to stabilize
      print('[WebSocket Test] Waiting for connection to stabilize...');
      await Future.delayed(const Duration(milliseconds: 1000));

      // Check if actually connected
      final service = ref.read(webSocketServiceProvider);
      final isConnected = service.isConnected;
      print('[WebSocket Test] Connection status after delay: $isConnected');

      // If we have a stored token, try to authenticate
      if (storedToken != null && storedToken.isNotEmpty && isConnected) {
        print(
          '[WebSocket Test] Attempting authentication with stored token...',
        );
        final result = await ref
            .read(webSocketControllerProvider.notifier)
            .authenticateWithStoredToken();
        print(
          '[WebSocket Test] Authentication with stored token result: $result',
        );
      } else if (!isConnected) {
        print('[WebSocket Test] Skipping authentication - not connected');
      } else {
        print('[WebSocket Test] No stored token, skipping auto-authentication');
      }
    } catch (e) {
      print('[WebSocket Test] Auto-connect failed: $e');
      log('Auto-connect failed: $e');
    }
  }

  @override
  void dispose() {
    _tokenController.dispose();
    _roomController.dispose();
    _messageController.dispose();
    _requestTestController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final connectionStatus = ref.watch(connectionStatusProvider);
    final joinedRooms = ref.watch(roomControllerProvider);
    final currentUser = ref.watch(currentUserProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('WebSocket Test'),
        backgroundColor: Theme.of(context).colorScheme.inversePrimary,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Connection status
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Connection Status',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Text('Status: $connectionStatus'),
                    if (currentUser != null) ...[
                      Text('User ID: ${currentUser['userID']}'),
                      Text('Username: ${currentUser['username']}'),
                    ],
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            await ref
                                .read(webSocketControllerProvider.notifier)
                                .connect();
                          },
                          child: const Text('Connect'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () async {
                            await ref
                                .read(webSocketControllerProvider.notifier)
                                .disconnect();
                          },
                          child: const Text('Disconnect'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () async {
                            await ref
                                .read(webSocketControllerProvider.notifier)
                                .reconnectWithAuth();
                          },
                          child: const Text('Reconnect'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Performance Metrics
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Performance Metrics',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Text('Total Events Received: ${_eventLog.length}'),
                    Text('Request/Response Tests: ${_requestResponses.length}'),
                    Text('Subscription Events: ${_subscriptionEvents.length}'),
                    Text('Current Rooms: ${joinedRooms.length}'),
                    if (currentUser != null)
                      Text('Authenticated as: ${currentUser['username']}'),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        ElevatedButton(
                          onPressed: () {
                            setState(() {
                              _eventLog.clear();
                              _requestResponses.clear();
                              _subscriptionEvents.clear();
                            });
                            _showSnackBar('All logs cleared');
                          },
                          child: const Text('Clear All Logs'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () {
                            final stats =
                                '''
Performance Summary:
- Total Events: ${_eventLog.length}
- Requests/Responses: ${_requestResponses.length}  
- Subscription Events: ${_subscriptionEvents.length}
- Current Status: $connectionStatus
- Joined Rooms: ${joinedRooms.join(', ')}
- User: ${currentUser?['username'] ?? 'Not authenticated'}
                            ''';

                            showDialog(
                              context: context,
                              builder: (context) => AlertDialog(
                                title: const Text('Connection Statistics'),
                                content: Text(stats),
                                actions: [
                                  TextButton(
                                    onPressed: () =>
                                        Navigator.of(context).pop(),
                                    child: const Text('Close'),
                                  ),
                                ],
                              ),
                            );
                          },
                          child: const Text('Show Stats'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Authentication
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Authentication',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: _tokenController,
                      decoration: const InputDecoration(
                        labelText: 'JWT Token',
                        hintText: 'Enter your JWT token',
                      ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            final token = _tokenController.text.trim();
                            print(
                              '[WebSocket Test] Authentication button pressed with token: $token',
                            );
                            if (token.isNotEmpty) {
                              print('[WebSocket Test] Calling authenticate...');
                              final success = await ref
                                  .read(webSocketControllerProvider.notifier)
                                  .authenticate(token);
                              print(
                                '[WebSocket Test] Authentication result: $success',
                              );
                              _showSnackBar(
                                success
                                    ? 'Authentication successful'
                                    : 'Authentication failed',
                              );
                            } else {
                              print(
                                '[WebSocket Test] Authentication skipped - empty token',
                              );
                            }
                          },
                          child: const Text('Authenticate'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            final success = await ref
                                .read(webSocketControllerProvider.notifier)
                                .authenticateWithStoredToken();
                            _showSnackBar(
                              success
                                  ? 'Authentication successful'
                                  : 'No stored token or authentication failed',
                            );
                          },
                          child: const Text('Use Stored Token'),
                        ),
                        ElevatedButton(
                          onPressed: () {
                            // Generate a simple test token for testing purposes
                            _tokenController.text = 'test-token-123';
                            _showSnackBar(
                              'Test token generated (for testing only)',
                            );
                          },
                          child: const Text('Generate Test Token'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Basic WebSocket Test
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Basic WebSocket Test',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    ElevatedButton(
                      onPressed: () async {
                        print('[WebSocket Test] Test Echo button pressed');
                        try {
                          final service = ref.read(webSocketServiceProvider);
                          print(
                            '[WebSocket Test] Sending test_echo request...',
                          );
                          final response = await service.sendRequest(
                            'test_echo',
                            'hello',
                          );
                          print(
                            '[WebSocket Test] test_echo response: success=${response.success}, data=${response.data}',
                          );
                          _showSnackBar(
                            'Test Echo: ${response.success ? 'SUCCESS' : 'FAILED'} - ${response.data}',
                          );
                        } catch (e) {
                          print('[WebSocket Test] Test Echo failed: $e');
                          _showSnackBar('Test Echo failed: $e');
                        }
                      },
                      child: const Text('Test Echo'),
                    ),
                    const SizedBox(width: 8),
                    ElevatedButton(
                      onPressed: () async {
                        print(
                          '[WebSocket Test] Test Create Room button pressed',
                        );
                        try {
                          final service = ref.read(webSocketServiceProvider);
                          print(
                            '[WebSocket Test] Sending create_room request...',
                          );
                          final response = await service.sendRequest(
                            'create_room',
                            'TestRoom123',
                          );
                          print(
                            '[WebSocket Test] create_room response: success=${response.success}, data=${response.data}, error=${response.error}',
                          );
                          _showSnackBar(
                            'Create Room: ${response.success ? 'SUCCESS' : 'FAILED'} - ${response.data ?? response.error}',
                          );
                        } catch (e) {
                          print('[WebSocket Test] Test Create Room failed: $e');
                          _showSnackBar('Test Create Room failed: $e');
                        }
                      },
                      child: const Text('Test Create Room'),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Room management
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Room Management',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Text('Joined Rooms: ${joinedRooms.join(', ')}'),
                    const SizedBox(height: 8),
                    TextField(
                      controller: _roomController,
                      decoration: const InputDecoration(
                        labelText: 'Room Name',
                        hintText: 'Enter room name',
                      ),
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            final room = _roomController.text.trim();
                            print(
                              '[WebSocket Test] Create Room button pressed with room: $room',
                            );
                            if (room.isNotEmpty) {
                              try {
                                final service = ref.read(
                                  webSocketServiceProvider,
                                );
                                print(
                                  '[WebSocket Test] Sending create_room message...',
                                );
                                // Send room name directly as data, not in a map
                                service.sendMessage('create_room', room);
                                print(
                                  '[WebSocket Test] create_room message sent',
                                );
                                _showSnackBar(
                                  'Room creation request sent: $room',
                                );
                              } catch (e) {
                                print(
                                  '[WebSocket Test] Error creating room: $e',
                                );
                                _showSnackBar('Failed to create room: $e');
                              }
                            } else {
                              print(
                                '[WebSocket Test] Create room skipped - empty room name',
                              );
                            }
                          },
                          child: const Text('Create Room'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            final room = _roomController.text.trim();
                            print(
                              '[WebSocket Test] Create & Join button pressed with room: $room',
                            );
                            if (room.isNotEmpty) {
                              try {
                                // First create the room
                                final service = ref.read(
                                  webSocketServiceProvider,
                                );
                                print(
                                  '[WebSocket Test] Sending create_room message for Create & Join...',
                                );

                                // Use sendRequest instead of sendMessage to wait for response
                                final createResponse = await service
                                    .sendRequest('create_room', room);
                                print(
                                  '[WebSocket Test] Create room response: success=${createResponse.success}, error=${createResponse.error}',
                                );

                                if (createResponse.success) {
                                  print(
                                    '[WebSocket Test] Room created successfully',
                                  );

                                  // Check if user was auto-joined during creation
                                  final data =
                                      createResponse.data
                                          as Map<String, dynamic>?;
                                  final autoJoined = data?['joined'] == true;

                                  if (autoJoined) {
                                    print(
                                      '[WebSocket Test] User was automatically joined to room during creation',
                                    );
                                    // Update room state manually since we were auto-joined
                                    ref
                                        .read(roomControllerProvider.notifier)
                                        .handleAutoJoin(room);
                                    _showSnackBar(
                                      'Room created and joined: $room',
                                    );
                                  } else {
                                    print(
                                      '[WebSocket Test] Need to manually join room...',
                                    );
                                    // Add a small delay to ensure server state is consistent
                                    await Future.delayed(
                                      Duration(milliseconds: 100),
                                    );

                                    print(
                                      '[WebSocket Test] About to call joinRoom($room)...',
                                    );
                                    final success = await ref
                                        .read(roomControllerProvider.notifier)
                                        .joinRoom(room);
                                    print(
                                      '[WebSocket Test] Join room result after creation: $success',
                                    );

                                    _showSnackBar(
                                      success
                                          ? 'Room created and joined: $room'
                                          : 'Room created but failed to join: $room',
                                    );
                                  }
                                } else {
                                  print(
                                    '[WebSocket Test] Room creation failed: ${createResponse.error}',
                                  );
                                  _showSnackBar(
                                    'Failed to create room: ${createResponse.error}',
                                  );
                                }
                              } catch (e) {
                                print(
                                  '[WebSocket Test] Error in create and join: $e',
                                );
                                _showSnackBar(
                                  'Failed to create and join room: $e',
                                );
                              }
                            } else {
                              print(
                                '[WebSocket Test] Create & Join skipped - empty room name',
                              );
                            }
                          },
                          child: const Text('Create & Join'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            final room = _roomController.text.trim();
                            print(
                              '[WebSocket Test] Join Room button pressed with room: $room',
                            );
                            if (room.isNotEmpty) {
                              print('[WebSocket Test] Calling joinRoom...');
                              final success = await ref
                                  .read(roomControllerProvider.notifier)
                                  .joinRoom(room);
                              print(
                                '[WebSocket Test] Join room result: $success',
                              );
                              _showSnackBar(
                                success
                                    ? 'Joined room: $room'
                                    : 'Failed to join room: $room',
                              );
                            } else {
                              print(
                                '[WebSocket Test] Join room skipped - empty room name',
                              );
                            }
                          },
                          child: const Text('Join Room'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            final room = _roomController.text.trim();
                            if (room.isNotEmpty) {
                              final success = await ref
                                  .read(roomControllerProvider.notifier)
                                  .leaveRoom(room);
                              _showSnackBar(
                                success
                                    ? 'Left room: $room'
                                    : 'Failed to leave room: $room',
                              );
                            }
                          },
                          child: const Text('Leave Room'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            print(
                              '[WebSocket Test] List All Rooms button pressed',
                            );
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              print(
                                '[WebSocket Test] Sending list_rooms message...',
                              );
                              // Send null/empty data for list rooms since it doesn't need data
                              service.sendMessage('list_rooms', null);
                              print('[WebSocket Test] list_rooms message sent');
                              _showSnackBar('Room list request sent');
                            } catch (e) {
                              print('[WebSocket Test] Error listing rooms: $e');
                              _showSnackBar('Failed to get room list: $e');
                            }
                          },
                          child: const Text('List All Rooms'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Message sending
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Send Message',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: _messageController,
                      decoration: const InputDecoration(
                        labelText: 'Message',
                        hintText: 'Enter message to send',
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            final message = _messageController.text.trim();
                            final room = _roomController.text.trim();
                            if (message.isNotEmpty && room.isNotEmpty) {
                              final success = await ref
                                  .read(roomControllerProvider.notifier)
                                  .sendToRoom(room, message);
                              _showSnackBar(
                                success
                                    ? 'Message sent'
                                    : 'Failed to send message',
                              );
                              if (success) {
                                _messageController.clear();
                              }
                            }
                          },
                          child: const Text('Send to Room'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () {
                            final service = ref.read(webSocketServiceProvider);
                            final message = _messageController.text.trim();
                            if (message.isNotEmpty) {
                              service.sendMessage('echo', message);
                              _showSnackBar('Echo message sent');
                              _messageController.clear();
                            }
                          },
                          child: const Text('Send Echo'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Request/Response Testing
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Request/Response Testing',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    TextField(
                      controller: _requestTestController,
                      decoration: const InputDecoration(
                        labelText: 'Test Message',
                        hintText: 'Enter message to test request/response',
                      ),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        ElevatedButton(
                          onPressed: _isAwaitingResponse
                              ? null
                              : () async {
                                  final message = _requestTestController.text
                                      .trim();
                                  if (message.isNotEmpty) {
                                    setState(() {
                                      _isAwaitingResponse = true;
                                      _lastRequestId = DateTime.now()
                                          .millisecondsSinceEpoch
                                          .toString();
                                    });

                                    try {
                                      final service = ref.read(
                                        webSocketServiceProvider,
                                      );
                                      final response = await service
                                          .sendRequest('test_request', {
                                            'id': _lastRequestId,
                                            'message': message,
                                            'timestamp': DateTime.now()
                                                .toIso8601String(),
                                          });

                                      setState(() {
                                        _isAwaitingResponse = false;
                                      });
                                      _logRequestResponse(
                                        'Request: $message → Response: $response',
                                      );
                                      _showSnackBar(
                                        'Request completed successfully',
                                      );
                                    } catch (e) {
                                      setState(() {
                                        _isAwaitingResponse = false;
                                      });
                                      _logRequestResponse(
                                        'Request: $message → Error: $e',
                                      );
                                      _showSnackBar('Request failed: $e');
                                    }
                                  }
                                },
                          child: _isAwaitingResponse
                              ? const SizedBox(
                                  width: 16,
                                  height: 16,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                )
                              : const Text('Send Request'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              final response = await service.sendRequest(
                                'ping',
                                {'timestamp': DateTime.now().toIso8601String()},
                              );

                              _logRequestResponse('Ping → Response: $response');
                              _showSnackBar('Ping successful');
                            } catch (e) {
                              _logRequestResponse('Ping → Error: $e');
                              _showSnackBar('Ping failed: $e');
                            }
                          },
                          child: const Text('Ping Server'),
                        ),
                        const SizedBox(width: 8),
                        ElevatedButton(
                          onPressed: () {
                            setState(() {
                              _requestResponses.clear();
                            });
                          },
                          child: const Text('Clear'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Container(
                      height: 120,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: ListView.builder(
                        itemCount: _requestResponses.length,
                        itemBuilder: (context, index) {
                          return Padding(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8.0,
                              vertical: 4.0,
                            ),
                            child: Text(
                              _requestResponses[index],
                              style: const TextStyle(fontSize: 12),
                            ),
                          );
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Subscription Testing
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Subscription & Broadcast Testing',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              service.sendMessage('subscribe', {
                                'event': 'server_announcements',
                              });
                              setState(() {
                                _subscriptionEvents.add(
                                  '[${DateTime.now().toIso8601String()}] Subscribed to server_announcements',
                                );
                              });
                              _showSnackBar(
                                'Subscribed to server announcements',
                              );
                            } catch (e) {
                              _showSnackBar('Subscription failed: $e');
                            }
                          },
                          child: const Text('Subscribe to Announcements'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              service.sendMessage('subscribe', {
                                'event': 'user_activity',
                              });
                              setState(() {
                                _subscriptionEvents.add(
                                  '[${DateTime.now().toIso8601String()}] Subscribed to user_activity',
                                );
                              });
                              _showSnackBar('Subscribed to user activity');
                            } catch (e) {
                              _showSnackBar('Subscription failed: $e');
                            }
                          },
                          child: const Text('Subscribe to User Activity'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              service.sendMessage('trigger_broadcast', {
                                'message':
                                    'Test broadcast from ${currentUser?['username'] ?? 'Anonymous'}',
                                'timestamp': DateTime.now().toIso8601String(),
                              });
                              _showSnackBar('Broadcast trigger sent');
                            } catch (e) {
                              _showSnackBar('Broadcast trigger failed: $e');
                            }
                          },
                          child: const Text('Trigger Test Broadcast'),
                        ),
                        ElevatedButton(
                          onPressed: () {
                            setState(() {
                              _subscriptionEvents.clear();
                            });
                          },
                          child: const Text('Clear'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Container(
                      height: 120,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: ListView.builder(
                        itemCount: _subscriptionEvents.length,
                        itemBuilder: (context, index) {
                          return Padding(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8.0,
                              vertical: 4.0,
                            ),
                            child: Text(
                              _subscriptionEvents[index],
                              style: const TextStyle(fontSize: 12),
                            ),
                          );
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Connection Testing
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Connection Stress Testing',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            _showSnackBar('Starting rapid connection test...');
                            for (int i = 0; i < 5; i++) {
                              await ref
                                  .read(webSocketControllerProvider.notifier)
                                  .disconnect();
                              await Future.delayed(
                                const Duration(milliseconds: 500),
                              );
                              await ref
                                  .read(webSocketControllerProvider.notifier)
                                  .connect();
                              await Future.delayed(
                                const Duration(milliseconds: 500),
                              );
                            }
                            _showSnackBar('Rapid connection test completed');
                          },
                          child: const Text('Rapid Connect/Disconnect'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            _showSnackBar('Sending message burst...');
                            final service = ref.read(webSocketServiceProvider);
                            for (int i = 0; i < 10; i++) {
                              service.sendMessage('burst_test', {
                                'index': i,
                                'timestamp': DateTime.now().toIso8601String(),
                              });
                              await Future.delayed(
                                const Duration(milliseconds: 100),
                              );
                            }
                            _showSnackBar('Message burst completed');
                          },
                          child: const Text('Message Burst Test'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            _showSnackBar('Starting heartbeat test...');
                            final service = ref.read(webSocketServiceProvider);
                            for (int i = 0; i < 5; i++) {
                              try {
                                final start = DateTime.now();
                                final response = await service.sendRequest(
                                  'heartbeat',
                                  {'ping': start.toIso8601String()},
                                );
                                final end = DateTime.now();
                                final latency = end
                                    .difference(start)
                                    .inMilliseconds;
                                _logSubscriptionEvent(
                                  'Heartbeat ${i + 1}: ${latency}ms - $response',
                                );
                                await Future.delayed(
                                  const Duration(seconds: 1),
                                );
                              } catch (e) {
                                _logSubscriptionEvent(
                                  'Heartbeat ${i + 1} failed: $e',
                                );
                              }
                            }
                            _showSnackBar('Heartbeat test completed');
                          },
                          child: const Text('Heartbeat Test'),
                        ),
                        ElevatedButton(
                          onPressed: _runConnectionTest,
                          child: const Text('Full Connection Test'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Error and Edge Case Testing
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Error & Edge Case Testing',
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              service.sendMessage('invalid_message_type', {
                                'data': 'This should trigger an error response',
                              });
                              _showSnackBar('Invalid message sent');
                            } catch (e) {
                              _showSnackBar(
                                'Error sending invalid message: $e',
                              );
                            }
                          },
                          child: const Text('Send Invalid Message'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              // Send malformed JSON-like data
                              service.sendMessage('test', {
                                'circular_ref': 'test',
                                'large_data': List.generate(
                                  1000,
                                  (i) => 'data_$i',
                                ),
                              });
                              _showSnackBar('Large payload sent');
                            } catch (e) {
                              _showSnackBar('Error sending large payload: $e');
                            }
                          },
                          child: const Text('Send Large Payload'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final service = ref.read(
                                webSocketServiceProvider,
                              );
                              // Try to send request that should timeout
                              final future = service.sendRequest(
                                'timeout_test',
                                {
                                  'delay': 30000, // 30 seconds
                                },
                              );

                              final response = await future.timeout(
                                const Duration(seconds: 5),
                                onTimeout: () => throw TimeoutException(
                                  'Request timed out',
                                  const Duration(seconds: 5),
                                ),
                              );

                              _logRequestResponse(
                                'Timeout test → Response: $response',
                              );
                            } catch (e) {
                              _logRequestResponse('Timeout test → Error: $e');
                              _showSnackBar('Timeout test completed');
                            }
                          },
                          child: const Text('Test Request Timeout'),
                        ),
                        ElevatedButton(
                          onPressed: () async {
                            try {
                              final controller = ref.read(
                                webSocketControllerProvider.notifier,
                              );
                              await controller.authenticate(
                                'invalid_token_12345',
                              );
                            } catch (e) {
                              _showSnackBar('Invalid auth test completed: $e');
                            }
                          },
                          child: const Text('Test Invalid Auth'),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Event log
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        Text(
                          'Event Log',
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        TextButton(
                          onPressed: () {
                            setState(() {
                              _eventLog.clear();
                            });
                          },
                          child: const Text('Clear'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Container(
                      height: 200,
                      decoration: BoxDecoration(
                        border: Border.all(color: Colors.grey),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: ListView.builder(
                        itemCount: _eventLog.length,
                        itemBuilder: (context, index) {
                          return Padding(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 8.0,
                              vertical: 4.0,
                            ),
                            child: Text(
                              _eventLog[index],
                              style: const TextStyle(fontSize: 12),
                            ),
                          );
                        },
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _logEvent(String message) {
    setState(() {
      _eventLog.add('[${DateTime.now().toIso8601String()}] $message');
    });
  }

  void _logSubscriptionEvent(String message) {
    setState(() {
      _subscriptionEvents.add('[${DateTime.now().toIso8601String()}] $message');
    });
  }

  void _logRequestResponse(String message) {
    setState(() {
      _requestResponses.add('[${DateTime.now().toIso8601String()}] $message');
    });
  }

  Future<void> _runConnectionTest() async {
    _showSnackBar('Running comprehensive connection test...');

    try {
      // Test disconnect/reconnect cycle
      _logEvent('Starting connection test');

      await ref.read(webSocketControllerProvider.notifier).disconnect();
      await Future.delayed(const Duration(seconds: 1));

      await ref.read(webSocketControllerProvider.notifier).connect();
      await Future.delayed(const Duration(seconds: 1));

      // Test authentication
      if (_tokenController.text.isNotEmpty) {
        final authSuccess = await ref
            .read(webSocketControllerProvider.notifier)
            .authenticate(_tokenController.text);
        _logEvent('Auth test: ${authSuccess ? 'SUCCESS' : 'FAILED'}');
      }

      // Test room operations
      if (_roomController.text.isNotEmpty) {
        final joinSuccess = await ref
            .read(roomControllerProvider.notifier)
            .joinRoom(_roomController.text);
        _logEvent('Room join test: ${joinSuccess ? 'SUCCESS' : 'FAILED'}');

        if (joinSuccess) {
          await Future.delayed(const Duration(milliseconds: 500));
          final leaveSuccess = await ref
              .read(roomControllerProvider.notifier)
              .leaveRoom(_roomController.text);
          _logEvent('Room leave test: ${leaveSuccess ? 'SUCCESS' : 'FAILED'}');
        }
      }

      _showSnackBar('Connection test completed');
    } catch (e) {
      _logEvent('Connection test error: $e');
      _showSnackBar('Connection test failed: $e');
    }
  }

  void _showSnackBar(String message) {
    ScaffoldMessenger.of(
      context,
    ).showSnackBar(SnackBar(content: Text(message)));
  }
}
