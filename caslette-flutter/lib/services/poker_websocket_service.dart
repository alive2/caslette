import 'dart:convert';
import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import '../models/poker/poker_models.dart';

enum PokerMessageType {
  joinTable,
  leaveTable,
  playerAction,
  gameState,
  tableList,
  createTable,
  tableCreated,
  tableJoined,
  tableLeft,
  gameStarted,
  handDealt,
  actionRequired,
  roundComplete,
  gameEnded,
  error;

  factory PokerMessageType.fromString(String type) {
    switch (type) {
      case 'poker_join_table':
        return PokerMessageType.joinTable;
      case 'poker_leave_table':
        return PokerMessageType.leaveTable;
      case 'poker_player_action':
        return PokerMessageType.playerAction;
      case 'poker_get_game_state':
      case 'poker_game_state':
        return PokerMessageType.gameState;
      case 'poker_list_tables':
        return PokerMessageType.tableList;
      case 'poker_create_table':
        return PokerMessageType.createTable;
      case 'poker_table_created':
        return PokerMessageType.tableCreated;
      case 'poker_table_joined':
        return PokerMessageType.tableJoined;
      case 'poker_table_left':
        return PokerMessageType.tableLeft;
      case 'poker_game_started':
        return PokerMessageType.gameStarted;
      case 'poker_hand_dealt':
        return PokerMessageType.handDealt;
      case 'poker_action_required':
        return PokerMessageType.actionRequired;
      case 'poker_round_complete':
        return PokerMessageType.roundComplete;
      case 'poker_game_ended':
        return PokerMessageType.gameEnded;
      case 'poker_error':
      case 'error':
        return PokerMessageType.error;
      default:
        return PokerMessageType.error;
    }
  }

  String get messageType {
    switch (this) {
      case PokerMessageType.joinTable:
        return 'poker_join_table';
      case PokerMessageType.leaveTable:
        return 'poker_leave_table';
      case PokerMessageType.playerAction:
        return 'poker_player_action';
      case PokerMessageType.gameState:
        return 'poker_get_game_state';
      case PokerMessageType.tableList:
        return 'poker_list_tables';
      case PokerMessageType.createTable:
        return 'poker_create_table';
      case PokerMessageType.tableCreated:
        return 'poker_table_created';
      case PokerMessageType.tableJoined:
        return 'poker_table_joined';
      case PokerMessageType.tableLeft:
        return 'poker_table_left';
      case PokerMessageType.gameStarted:
        return 'poker_game_started';
      case PokerMessageType.handDealt:
        return 'poker_hand_dealt';
      case PokerMessageType.actionRequired:
        return 'poker_action_required';
      case PokerMessageType.roundComplete:
        return 'poker_round_complete';
      case PokerMessageType.gameEnded:
        return 'poker_game_ended';
      case PokerMessageType.error:
        return 'poker_error';
    }
  }
}

class PokerMessage {
  final PokerMessageType type;
  final Map<String, dynamic> data;
  final String? tableId;
  final String? error;

  const PokerMessage({
    required this.type,
    this.data = const {},
    this.tableId,
    this.error,
  });

  factory PokerMessage.fromJson(Map<String, dynamic> json) {
    return PokerMessage(
      type: PokerMessageType.fromString(json['type'] ?? ''),
      data: json['data'] ?? {},
      tableId: json['table_id'],
      error: json['error'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'type': type.messageType,
      'data': data,
      if (tableId != null) 'table_id': tableId,
      if (error != null) 'error': error,
    };
  }
}

class PokerWebSocketService extends ChangeNotifier {
  WebSocketChannel? _channel;
  final StreamController<PokerMessage> _messageController =
      StreamController<PokerMessage>.broadcast();
  bool _isConnected = false;
  String? _token;
  bool _isDisposed = false;

  // Getters
  bool get isConnected => _isConnected;
  Stream<PokerMessage> get messages => _messageController.stream;

  // Connect to poker WebSocket
  Future<void> connect(String token) async {
    if (_isConnected || _isDisposed) {
      debugPrint(
        'Poker WebSocket already connected or disposed - isConnected: $_isConnected, isDisposed: $_isDisposed',
      );
      return;
    }

    debugPrint('Starting poker WebSocket connection...');
    try {
      _token = token;
      final uri = Uri.parse('ws://localhost:8080/api/v1/ws?token=$token');
      _channel = WebSocketChannel.connect(uri);

      // Set up listeners before marking as connected
      _channel!.stream.listen(
        _handleMessage,
        onError: _handleError,
        onDone: _handleDisconnection,
      );

      // Wait for the first message to confirm connection is ready
      debugPrint('Waiting for WebSocket to be ready...');
      await _channel!.ready;

      _isConnected = true;
      debugPrint('Poker WebSocket connected successfully');
      if (!_isDisposed) {
        debugPrint('Notifying listeners of connection change...');
        notifyListeners();
      }
    } catch (e) {
      debugPrint('Poker WebSocket connection failed: $e');
      _handleError(e);
    }
  }

  // Disconnect from WebSocket
  void disconnect() {
    if (_channel != null) {
      _channel!.sink.close();
      _channel = null;
    }
    _isConnected = false;
    debugPrint('Poker WebSocket disconnected');
    if (!_isDisposed) {
      notifyListeners();
    }
  }

  // Send a poker message
  void sendMessage(PokerMessage message) {
    if (!_isConnected || _channel == null) {
      debugPrint('Cannot send message: WebSocket not connected');
      return;
    }

    try {
      // Send message directly in the format expected by Go server
      final messageData = {
        'type': message.type.messageType,
        'data': message.data,
        if (message.tableId != null) 'table_id': message.tableId,
        if (message.error != null) 'error': message.error,
      };

      final jsonMessage = jsonEncode(messageData);
      _channel!.sink.add(jsonMessage);
      debugPrint('Sent poker message: ${message.type.messageType}');
    } catch (e) {
      debugPrint('Failed to send poker message: $e');
    }
  }

  // Request list of available tables
  void requestTableList() {
    sendMessage(const PokerMessage(type: PokerMessageType.tableList));
  }

  // Create a new poker table
  void createTable({
    required String name,
    required int minBuyIn,
    required int maxBuyIn,
    required int smallBlind,
    required int bigBlind,
    required int maxPlayers,
  }) {
    sendMessage(
      PokerMessage(
        type: PokerMessageType.createTable,
        data: {
          'name': name,
          'min_buy_in': minBuyIn,
          'max_buy_in': maxBuyIn,
          'small_blind': smallBlind,
          'big_blind': bigBlind,
          'max_players': maxPlayers,
          'game_type': 'texas_holdem',
          'rake_percent': 0.05,
          'max_rake': 50,
          'is_private': false,
          'password': '',
        },
      ),
    );
  }

  // Join a poker table
  void joinTable(String tableId, int seatNumber, int buyInAmount) {
    sendMessage(
      PokerMessage(
        type: PokerMessageType.joinTable,
        tableId: tableId,
        data: {'seat_number': seatNumber, 'buy_in_amount': buyInAmount},
      ),
    );
  }

  // Leave a poker table
  void leaveTable(String tableId) {
    sendMessage(
      PokerMessage(type: PokerMessageType.leaveTable, tableId: tableId),
    );
  }

  // Perform a player action
  void performAction(String tableId, PlayerAction action, {int? amount}) {
    sendMessage(
      PokerMessage(
        type: PokerMessageType.playerAction,
        tableId: tableId,
        data: {'action': action.name, if (amount != null) 'amount': amount},
      ),
    );
  }

  // Request current game state
  void requestGameState(String tableId) {
    sendMessage(
      PokerMessage(type: PokerMessageType.gameState, tableId: tableId),
    );
  }

  // Handle incoming messages
  void _handleMessage(dynamic data) {
    try {
      final Map<String, dynamic> messageData = jsonDecode(data);

      // Check if this is a poker message (starts with 'poker_')
      final messageType = messageData['type'] as String?;
      if (messageType != null && messageType.startsWith('poker_')) {
        final pokerMessage = PokerMessage.fromJson(messageData);
        debugPrint('Received poker message: ${pokerMessage.type.messageType}');
        _messageController.add(pokerMessage);
      }
    } catch (e) {
      debugPrint('Failed to parse poker message: $e');
    }
  }

  // Handle WebSocket errors
  void _handleError(dynamic error) {
    debugPrint('Poker WebSocket error: $error');
    _isConnected = false;
    if (!_isDisposed) {
      notifyListeners();

      // Emit error message
      _messageController.add(
        PokerMessage(type: PokerMessageType.error, error: error.toString()),
      );
    }
  }

  // Handle WebSocket disconnection
  void _handleDisconnection() {
    debugPrint('Poker WebSocket disconnected');
    _isConnected = false;
    if (!_isDisposed) {
      notifyListeners();

      // Attempt to reconnect after a delay only if not disposed
      if (_token != null) {
        Future.delayed(const Duration(seconds: 5), () {
          if (!_isConnected && !_isDisposed) {
            debugPrint('Attempting to reconnect poker WebSocket...');
            connect(_token!);
          }
        });
      }
    }
  }

  @override
  void dispose() {
    _isDisposed = true;
    disconnect();
    _messageController.close();
    super.dispose();
  }
}
