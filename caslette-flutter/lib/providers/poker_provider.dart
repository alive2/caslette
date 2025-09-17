import 'dart:async';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter/foundation.dart';
import '../models/poker/poker_models.dart';
import '../services/poker_websocket_service.dart';

// Poker WebSocket Service Provider
final pokerWebSocketServiceProvider =
    ChangeNotifierProvider<PokerWebSocketService>((ref) {
      return PokerWebSocketService();
    });

// Poker State Classes
class PokerState {
  final List<PokerTable> availableTables;
  final PokerTable? currentTable;
  final String? currentTableId;
  final bool isLoading;
  final String? error;
  final bool isConnected;

  const PokerState({
    this.availableTables = const [],
    this.currentTable,
    this.currentTableId,
    this.isLoading = false,
    this.error,
    this.isConnected = false,
  });

  PokerState copyWith({
    List<PokerTable>? availableTables,
    PokerTable? currentTable,
    String? currentTableId,
    bool? isLoading,
    String? error,
    bool? isConnected,
  }) {
    return PokerState(
      availableTables: availableTables ?? this.availableTables,
      currentTable: currentTable ?? this.currentTable,
      currentTableId: currentTableId ?? this.currentTableId,
      isLoading: isLoading ?? this.isLoading,
      error: error ?? this.error,
      isConnected: isConnected ?? this.isConnected,
    );
  }

  PokerState clearError() {
    return copyWith(error: null);
  }

  PokerState setError(String errorMessage) {
    return copyWith(error: errorMessage, isLoading: false);
  }
}

// Poker State Notifier
class PokerNotifier extends StateNotifier<PokerState> {
  final PokerWebSocketService _webSocketService;
  late final StreamSubscription _messageSubscription;
  bool _isDisposed = false;

  PokerNotifier(this._webSocketService) : super(const PokerState()) {
    _initializeMessageListener();
    _webSocketService.addListener(_onConnectionChange);
  }

  void _initializeMessageListener() {
    _messageSubscription = _webSocketService.messages.listen((message) {
      _handlePokerMessage(message);
    });
  }

  void _onConnectionChange() {
    if (!_isDisposed) {
      final wasConnected = state.isConnected;
      final isNowConnected = _webSocketService.isConnected;

      state = state.copyWith(isConnected: isNowConnected);

      // Auto-request table list when connection is established
      if (!wasConnected && isNowConnected) {
        Future.delayed(const Duration(milliseconds: 200), () {
          if (!_isDisposed && _webSocketService.isConnected) {
            requestTableList();
          }
        });
      }
    }
  }

  // Connect to poker WebSocket
  Future<void> connect(String token) async {
    if (_isDisposed) return;
    state = state.copyWith(isLoading: true, error: null);
    try {
      await _webSocketService.connect(token);
      if (!_isDisposed) {
        state = state.copyWith(isLoading: false, isConnected: true);
      }
    } catch (e) {
      if (!_isDisposed) {
        state = state.setError('Failed to connect: ${e.toString()}');
      }
    }
  }

  // Disconnect from poker WebSocket
  void disconnect() {
    _webSocketService.disconnect();
    if (!_isDisposed) {
      state = state.copyWith(
        isConnected: false,
        currentTable: null,
        currentTableId: null,
      );
    }
  }

  // Request list of available tables
  void requestTableList() {
    if (_isDisposed) return;

    // Check if WebSocket is connected
    if (!_webSocketService.isConnected) {
      state = state.setError('Not connected to poker server');
      return;
    }

    state = state.copyWith(isLoading: true, error: null);
    _webSocketService.requestTableList();
  }

  // Join a poker table
  void joinTable(String tableId, int seatNumber, int buyInAmount) {
    if (_isDisposed) return;
    state = state.copyWith(isLoading: true, error: null);
    _webSocketService.joinTable(tableId, seatNumber, buyInAmount);
  }

  // Leave current table
  void leaveTable() {
    if (state.currentTableId != null) {
      _webSocketService.leaveTable(state.currentTableId!);
      if (!_isDisposed) {
        state = state.copyWith(currentTable: null, currentTableId: null);
      }
    }
  }

  // Perform player action
  void performAction(PlayerAction action, {int? amount}) {
    if (state.currentTableId != null) {
      _webSocketService.performAction(
        state.currentTableId!,
        action,
        amount: amount,
      );
    }
  }

  // Request current game state
  void requestGameState() {
    if (state.currentTableId != null) {
      _webSocketService.requestGameState(state.currentTableId!);
    }
  }

  // Handle incoming poker messages
  void _handlePokerMessage(PokerMessage message) {
    if (_isDisposed) return;
    debugPrint('Handling poker message: ${message.type.messageType}');

    switch (message.type) {
      case PokerMessageType.tableList:
        _handleTableList(message);
        break;
      case PokerMessageType.tableJoined:
        _handleTableJoined(message);
        break;
      case PokerMessageType.tableLeft:
        _handleTableLeft(message);
        break;
      case PokerMessageType.gameState:
        _handleGameState(message);
        break;
      case PokerMessageType.gameStarted:
        _handleGameStarted(message);
        break;
      case PokerMessageType.handDealt:
        _handleHandDealt(message);
        break;
      case PokerMessageType.actionRequired:
        _handleActionRequired(message);
        break;
      case PokerMessageType.roundComplete:
        _handleRoundComplete(message);
        break;
      case PokerMessageType.gameEnded:
        _handleGameEnded(message);
        break;
      case PokerMessageType.error:
        _handleError(message);
        break;
      default:
        debugPrint('Unhandled poker message type: ${message.type.messageType}');
    }
  }

  void _handleTableList(PokerMessage message) {
    if (_isDisposed) return;
    try {
      final tablesData = message.data['tables'] as List<dynamic>? ?? [];
      final tables = tablesData
          .map((tableData) => PokerTable.fromJson(tableData))
          .toList();

      state = state.copyWith(
        availableTables: tables,
        isLoading: false,
        error: null,
      );
    } catch (e) {
      state = state.setError('Failed to parse table list: ${e.toString()}');
    }
  }

  void _handleTableJoined(PokerMessage message) {
    if (_isDisposed) return;
    try {
      final tableData = message.data['table'];
      if (tableData != null) {
        final table = PokerTable.fromJson(tableData);
        state = state.copyWith(
          currentTable: table,
          currentTableId: table.id,
          isLoading: false,
          error: null,
        );
      }
    } catch (e) {
      state = state.setError('Failed to join table: ${e.toString()}');
    }
  }

  void _handleTableLeft(PokerMessage message) {
    if (_isDisposed) return;
    state = state.copyWith(currentTable: null, currentTableId: null);
  }

  void _handleGameState(PokerMessage message) {
    if (_isDisposed) return;
    try {
      final tableData = message.data['table'];
      if (tableData != null) {
        final table = PokerTable.fromJson(tableData);
        state = state.copyWith(
          currentTable: table,
          isLoading: false,
          error: null,
        );
      }
    } catch (e) {
      state = state.setError('Failed to update game state: ${e.toString()}');
    }
  }

  void _handleGameStarted(PokerMessage message) {
    if (_isDisposed) return;
    // Game has started, request updated game state
    requestGameState();
  }

  void _handleHandDealt(PokerMessage message) {
    if (_isDisposed) return;
    // Hand has been dealt, request updated game state
    requestGameState();
  }

  void _handleActionRequired(PokerMessage message) {
    if (_isDisposed) return;
    // Action is required from player, request updated game state
    requestGameState();
  }

  void _handleRoundComplete(PokerMessage message) {
    if (_isDisposed) return;
    // Round is complete, request updated game state
    requestGameState();
  }

  void _handleGameEnded(PokerMessage message) {
    if (_isDisposed) return;
    // Game has ended, request updated game state
    requestGameState();
  }

  void _handleError(PokerMessage message) {
    if (_isDisposed) return;
    state = state.setError(message.error ?? 'Unknown poker error');
  }

  // Clear error state
  void clearError() {
    if (_isDisposed) return;
    state = state.clearError();
  }

  @override
  void dispose() {
    _isDisposed = true;
    _messageSubscription.cancel();
    _webSocketService.removeListener(_onConnectionChange);
    // Don't dispose the WebSocket service directly since it's managed by Riverpod
    // _webSocketService.dispose(); // This was causing the circular dependency
    super.dispose();
  }
}

// Poker Provider
final pokerProvider = StateNotifierProvider<PokerNotifier, PokerState>((ref) {
  final webSocketService = ref.watch(pokerWebSocketServiceProvider);
  return PokerNotifier(webSocketService);
});

// Convenience providers for specific poker state aspects
final availableTablesProvider = Provider<List<PokerTable>>((ref) {
  return ref.watch(pokerProvider).availableTables;
});

final currentTableProvider = Provider<PokerTable?>((ref) {
  return ref.watch(pokerProvider).currentTable;
});

final pokerConnectionProvider = Provider<bool>((ref) {
  return ref.watch(pokerProvider).isConnected;
});

final pokerLoadingProvider = Provider<bool>((ref) {
  return ref.watch(pokerProvider).isLoading;
});

final pokerErrorProvider = Provider<String?>((ref) {
  return ref.watch(pokerProvider).error;
});
