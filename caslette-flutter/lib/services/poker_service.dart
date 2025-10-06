import 'dart:async';
import 'dart:developer';
import '../models/poker_models.dart';
import 'websocket/websocket_service.dart';

class PokerService {
  final WebSocketService _webSocketService;

  // Stream controllers for poker events
  final StreamController<List<PokerTable>> _tablesController =
      StreamController<List<PokerTable>>.broadcast();
  final StreamController<PokerTable?> _currentTableController =
      StreamController<PokerTable?>.broadcast();
  final StreamController<PokerGameState?> _gameStateController =
      StreamController<PokerGameState?>.broadcast();
  final StreamController<String> _pokerErrorController =
      StreamController<String>.broadcast();
  final StreamController<Map<String, dynamic>> _pokerEventController =
      StreamController<Map<String, dynamic>>.broadcast();

  // Current state
  List<PokerTable> _tables = [];
  PokerTable? _currentTable;
  PokerGameState? _currentGameState;
  bool _isInTable = false;
  String? _currentTableRoom;

  PokerService(this._webSocketService) {
    _setupWebSocketListeners();
  }

  // Getters for streams
  Stream<List<PokerTable>> get tablesStream => _tablesController.stream;
  Stream<PokerTable?> get currentTableStream => _currentTableController.stream;
  Stream<PokerGameState?> get gameStateStream => _gameStateController.stream;
  Stream<String> get errorStream => _pokerErrorController.stream;
  Stream<Map<String, dynamic>> get eventStream => _pokerEventController.stream;

  // Getters for current state
  List<PokerTable> get tables => _tables;
  PokerTable? get currentTable => _currentTable;
  PokerGameState? get currentGameState => _currentGameState;
  bool get isInTable => _isInTable;

  void _setupWebSocketListeners() {
    // Listen for WebSocket events
    _webSocketService.eventStream.listen((event) {
      _handleWebSocketEvent(event);
    });
  }

  void _handleWebSocketEvent(WebSocketEvent event) {
    try {
      switch (event.type) {
        // Real-time events
        case 'player_joined':
        case 'player_left':
        case 'game_started':
        case 'game_ended':
        case 'poker_action_broadcast':
        case 'game_state_update':
        case 'table_updated':
          _handlePokerEvent(event);
          break;

        default:
          // Ignore other event types
          break;
      }
    } catch (e) {
      log('Error handling poker WebSocket event: $e');
      _pokerErrorController.add('Error processing event: $e');
    }
  }

  void _handlePokerEvent(WebSocketEvent event) {
    // Add to event stream for UI to handle
    _pokerEventController.add({'type': event.type, 'data': event.data});

    // Update local state based on event
    switch (event.type) {
      case 'game_state_update':
        if (event.data != null && event.data['game_state'] != null) {
          final PokerGameState gameState = PokerGameState.fromJson(
            event.data['game_state'] as Map<String, dynamic>,
          );
          _currentGameState = gameState;
          _gameStateController.add(_currentGameState);
        }
        break;
      case 'table_updated':
        if (event.data != null && event.data['table'] != null) {
          final PokerTable table = PokerTable.fromJson(
            event.data['table'] as Map<String, dynamic>,
          );
          _currentTable = table;
          _currentTableController.add(_currentTable);
        }
        break;
    }
  }

  // Public API methods

  Future<void> getTableList() async {
    try {
      final response = await _webSocketService.sendRequest('table_list', {});
      print('Table list response: ${response.success}, data: ${response.data}');

      if (response.success) {
        List<PokerTable> tables = [];

        // Handle different response formats
        if (response.data != null) {
          if (response.data is List) {
            // Backend returns data as a direct list of tables
            final List<dynamic> tablesData = response.data as List<dynamic>;
            tables = tablesData
                .map(
                  (tableData) =>
                      PokerTable.fromJson(tableData as Map<String, dynamic>),
                )
                .toList();
          } else if (response.data is Map<String, dynamic>) {
            final dataMap = response.data as Map<String, dynamic>;
            if (dataMap.containsKey('tables') && dataMap['tables'] is List) {
              final List<dynamic> tablesData =
                  dataMap['tables'] as List<dynamic>;
              tables = tablesData
                  .map(
                    (tableData) =>
                        PokerTable.fromJson(tableData as Map<String, dynamic>),
                  )
                  .toList();
            }
          }
        }

        _tables = tables;
        _tablesController.add(_tables);
        print('Loaded ${tables.length} tables');
      } else {
        print('Table list request failed: ${response.error}');
        _pokerErrorController.add(response.error ?? 'Failed to get table list');
      }
    } catch (e) {
      print('Error getting table list: $e');
      // For now, just return empty list instead of error for debugging
      _tables = [];
      _tablesController.add(_tables);
      print('Setting empty table list due to error');
    }
  }

  Future<PokerTable?> createTable({
    required String name,
    String gameType = 'texas_holdem',
    int maxPlayers = 8,
    int smallBlind = 10,
    int bigBlind = 20,
    bool autoStart = true,
    bool isPrivate = false,
    String? password,
  }) async {
    try {
      final response = await _webSocketService.sendRequest('table_create', {
        'name': name,
        'game_type': gameType,
        'max_players': maxPlayers,
        'settings': {
          'small_blind': smallBlind,
          'big_blind': bigBlind,
          'auto_start': autoStart,
          'buy_in': bigBlind * 20,
          'max_buy_in': bigBlind * 100,
          'time_limit': 30,
        },
        'is_private': isPrivate,
        if (password != null) 'password': password,
      });

      if (response.success && response.data != null) {
        final PokerTable table = PokerTable.fromJson(
          response.data as Map<String, dynamic>,
        );
        _tables.add(table);
        _tablesController.add(_tables);

        // Since the user is automatically joined when creating a table,
        // set this as the current table
        _currentTable = table;
        _isInTable = true;
        _currentTableController.add(_currentTable);

        // Join the table room for real-time updates
        await _joinTableRoom(table.id);

        return table;
      } else {
        _pokerErrorController.add(response.error ?? 'Failed to create table');
        return null;
      }
    } catch (e) {
      _pokerErrorController.add('Error creating table: $e');
      return null;
    }
  }

  Future<void> joinTable(
    String tableId, {
    String mode = 'player',
    String? password,
  }) async {
    try {
      final response = await _webSocketService.sendRequest('table_join', {
        'table_id': tableId,
        'mode': mode,
        if (password != null) 'password': password,
      });

      if (response.success && response.data != null) {
        final PokerTable table = PokerTable.fromJson(
          response.data['table'] as Map<String, dynamic>,
        );
        _currentTable = table;
        _isInTable = true;
        _currentTableController.add(_currentTable);

        // Join the table room for real-time updates
        await _joinTableRoom(table.id);
      } else {
        _pokerErrorController.add(response.error ?? 'Failed to join table');
      }
    } catch (e) {
      _pokerErrorController.add('Error joining table: $e');
    }
  }

  Future<void> leaveTable() async {
    if (_currentTable == null) {
      print('No current table to leave');
      return;
    }

    try {
      print('Attempting to leave table: ${_currentTable!.id}');
      print('WebSocket connected: ${_webSocketService.isConnected}');

      final response = await _webSocketService
          .sendRequest('table_leave', {'table_id': _currentTable!.id})
          .timeout(
            const Duration(seconds: 10),
            onTimeout: () {
              throw Exception(
                'Request timeout - server did not respond within 10 seconds',
              );
            },
          );

      print('Leave table response received:');
      print('- Success: ${response.success}');
      print('- Type: ${response.type}');
      print('- Data: ${response.data}');
      print('- Error: ${response.error}');

      if (response.success) {
        _currentTable = null;
        _currentGameState = null;
        _isInTable = false;
        _currentTableController.add(null);
        _gameStateController.add(null);

        // Leave the table room
        if (_currentTableRoom != null) {
          print('Leaving table room: $_currentTableRoom');
          await _leaveTableRoom(_currentTableRoom!);
          _currentTableRoom = null;
        }

        print('Successfully left table');
      } else {
        final errorMsg = response.error ?? 'Failed to leave table';
        print('Failed to leave table: $errorMsg');
        _pokerErrorController.add(errorMsg);
      }
    } catch (e) {
      print('Error leaving table: $e');
      _pokerErrorController.add('Error leaving table: $e');
    }
  }

  Future<void> setReady(bool ready) async {
    if (_currentTable == null) return;

    try {
      final response = await _webSocketService.sendRequest('table_set_ready', {
        'table_id': _currentTable!.id,
        'ready': ready,
      });

      if (!response.success) {
        _pokerErrorController.add(
          response.error ?? 'Failed to set ready status',
        );
      }
    } catch (e) {
      _pokerErrorController.add('Error setting ready status: $e');
    }
  }

  Future<void> startGame() async {
    if (_currentTable == null) return;

    try {
      final response = await _webSocketService.sendRequest('table_start_game', {
        'table_id': _currentTable!.id,
      });

      if (!response.success) {
        _pokerErrorController.add(response.error ?? 'Failed to start game');
      }
    } catch (e) {
      _pokerErrorController.add('Error starting game: $e');
    }
  }

  Future<void> getGameState() async {
    if (_currentTable == null) return;

    try {
      final response = await _webSocketService.sendRequest(
        'table_get_game_state',
        {'table_id': _currentTable!.id},
      );

      if (response.success && response.data != null) {
        final PokerGameState gameState = PokerGameState.fromJson(
          response.data['game_state'] as Map<String, dynamic>,
        );
        _currentGameState = gameState;
        _gameStateController.add(_currentGameState);
      } else {
        _pokerErrorController.add(response.error ?? 'Failed to get game state');
      }
    } catch (e) {
      _pokerErrorController.add('Error getting game state: $e');
    }
  }

  Future<void> performAction(PokerActionType action, {int? amount}) async {
    if (_currentTable == null) return;

    try {
      final data = <String, dynamic>{
        'table_id': _currentTable!.id,
        'action': action.name,
      };

      if (amount != null) {
        data['amount'] = amount;
      }

      final response = await _webSocketService.sendRequest(
        'poker_action',
        data,
      );

      if (!response.success) {
        _pokerErrorController.add(
          response.error ?? 'Failed to perform poker action',
        );
      }
    } catch (e) {
      _pokerErrorController.add('Error performing poker action: $e');
    }
  }

  Future<void> fold() async => performAction(PokerActionType.fold);
  Future<void> check() async => performAction(PokerActionType.check);
  Future<void> call() async => performAction(PokerActionType.call);
  Future<void> bet(int amount) async =>
      performAction(PokerActionType.bet, amount: amount);
  Future<void> raise(int amount) async =>
      performAction(PokerActionType.raise, amount: amount);
  Future<void> allIn() async => performAction(PokerActionType.allIn);

  Future<void> _joinTableRoom(String tableId) async {
    try {
      _currentTableRoom = 'table_$tableId';
      final response = await _webSocketService.sendRequest('join_table_room', {
        'table_id': tableId,
      });

      if (!response.success) {
        log('Failed to join table room: ${response.error}');
      }
    } catch (e) {
      log('Error joining table room: $e');
    }
  }

  Future<void> _leaveTableRoom(String room) async {
    try {
      await _webSocketService.leaveRoom(room);
    } catch (e) {
      log('Error leaving table room: $e');
    }
  }

  // Helper methods
  bool canPerformAction(PokerActionType action) {
    if (_currentGameState == null || _currentTable == null) return false;

    final myPlayer = _currentGameState!.getPlayerById(
      _webSocketService.connectionInfo.userID ?? '',
    );
    if (myPlayer == null) return false;

    // Check if it's my turn
    if (_currentGameState!.activePlayerID != myPlayer.id) return false;

    // Check if I'm still in the game
    if (myPlayer.isFolded || myPlayer.isAllIn) return false;

    // Action-specific checks
    switch (action) {
      case PokerActionType.fold:
        return true; // Can always fold
      case PokerActionType.check:
        return _currentGameState!.currentBet == myPlayer.currentBet;
      case PokerActionType.call:
        return _currentGameState!.currentBet > myPlayer.currentBet;
      case PokerActionType.bet:
        return _currentGameState!.currentBet == 0;
      case PokerActionType.raise:
        return _currentGameState!.currentBet > 0;
      case PokerActionType.allIn:
        return myPlayer.chips > 0;
    }
  }

  int getMinRaiseAmount() {
    if (_currentGameState == null) return 0;
    return _currentGameState!.currentBet * 2;
  }

  int getCallAmount() {
    if (_currentGameState == null || _currentTable == null) return 0;

    final myPlayer = _currentGameState!.getPlayerById(
      _webSocketService.connectionInfo.userID ?? '',
    );
    if (myPlayer == null) return 0;

    return _currentGameState!.currentBet - myPlayer.currentBet;
  }

  void dispose() {
    _tablesController.close();
    _currentTableController.close();
    _gameStateController.close();
    _pokerErrorController.close();
    _pokerEventController.close();
  }
}
