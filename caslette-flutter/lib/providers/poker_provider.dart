import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/poker_models.dart';
import '../services/poker_service.dart';
import '../services/websocket/websocket_providers.dart';

// Provider for poker service
final pokerServiceProvider = Provider<PokerService>((ref) {
  final webSocketService = ref.watch(webSocketServiceProvider);
  return PokerService(webSocketService);
});

// State notifier for poker tables
class PokerTablesNotifier extends StateNotifier<AsyncValue<List<PokerTable>>> {
  final PokerService _pokerService;

  PokerTablesNotifier(this._pokerService) : super(const AsyncValue.loading()) {
    _init();
  }

  void _init() {
    // Listen to tables stream
    _pokerService.tablesStream.listen(
      (tables) {
        state = AsyncValue.data(tables);
      },
      onError: (error) {
        state = AsyncValue.error(error, StackTrace.current);
      },
    );

    // Listen to error stream
    _pokerService.errorStream.listen((error) {
      state = AsyncValue.error(error, StackTrace.current);
    });

    // Load initial data
    refresh();
  }

  Future<void> refresh() async {
    state = const AsyncValue.loading();
    try {
      await _pokerService.getTableList();
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
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
      final table = await _pokerService.createTable(
        name: name,
        gameType: gameType,
        maxPlayers: maxPlayers,
        smallBlind: smallBlind,
        bigBlind: bigBlind,
        autoStart: autoStart,
        isPrivate: isPrivate,
        password: password,
      );
      return table;
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
      return null;
    }
  }
}

// Provider for poker tables
final pokerTablesProvider =
    StateNotifierProvider<PokerTablesNotifier, AsyncValue<List<PokerTable>>>((
      ref,
    ) {
      final pokerService = ref.watch(pokerServiceProvider);
      return PokerTablesNotifier(pokerService);
    });

// State notifier for current table
class CurrentTableNotifier extends StateNotifier<AsyncValue<PokerTable?>> {
  final PokerService _pokerService;

  CurrentTableNotifier(this._pokerService)
    : super(const AsyncValue.data(null)) {
    _init();
  }

  void _init() {
    // Listen to current table stream
    _pokerService.currentTableStream.listen(
      (table) {
        state = AsyncValue.data(table);
      },
      onError: (error) {
        state = AsyncValue.error(error, StackTrace.current);
      },
    );
  }

  Future<void> joinTable(
    String tableId, {
    String mode = 'player',
    String? password,
  }) async {
    try {
      await _pokerService.joinTable(tableId, mode: mode, password: password);
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  Future<void> leaveTable() async {
    try {
      await _pokerService.leaveTable();
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  Future<void> setReady(bool ready) async {
    try {
      await _pokerService.setReady(ready);
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  Future<void> startGame() async {
    try {
      await _pokerService.startGame();
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }
}

// Provider for current table
final currentTableProvider =
    StateNotifierProvider<CurrentTableNotifier, AsyncValue<PokerTable?>>((ref) {
      final pokerService = ref.watch(pokerServiceProvider);
      return CurrentTableNotifier(pokerService);
    });

// State notifier for game state
class GameStateNotifier extends StateNotifier<AsyncValue<PokerGameState?>> {
  final PokerService _pokerService;

  GameStateNotifier(this._pokerService) : super(const AsyncValue.data(null)) {
    _init();
  }

  void _init() {
    // Listen to game state stream
    _pokerService.gameStateStream.listen(
      (gameState) {
        state = AsyncValue.data(gameState);
      },
      onError: (error) {
        state = AsyncValue.error(error, StackTrace.current);
      },
    );
  }

  Future<void> refresh() async {
    try {
      await _pokerService.getGameState();
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
    }
  }

  Future<void> performAction(PokerActionType action, {int? amount}) async {
    try {
      await _pokerService.performAction(action, amount: amount);
    } catch (error, stackTrace) {
      state = AsyncValue.error(error, stackTrace);
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

  // Helper methods
  bool canPerformAction(PokerActionType action) {
    return _pokerService.canPerformAction(action);
  }

  int getMinRaiseAmount() {
    return _pokerService.getMinRaiseAmount();
  }

  int getCallAmount() {
    return _pokerService.getCallAmount();
  }
}

// Provider for game state
final gameStateProvider =
    StateNotifierProvider<GameStateNotifier, AsyncValue<PokerGameState?>>((
      ref,
    ) {
      final pokerService = ref.watch(pokerServiceProvider);
      return GameStateNotifier(pokerService);
    });

// Provider for poker events
final pokerEventsProvider = StreamProvider<Map<String, dynamic>>((ref) {
  final pokerService = ref.watch(pokerServiceProvider);
  return pokerService.eventStream;
});

// Provider for poker errors
final pokerErrorsProvider = StreamProvider<String>((ref) {
  final pokerService = ref.watch(pokerServiceProvider);
  return pokerService.errorStream;
});

// Provider for checking if user is in a table
final isInTableProvider = Provider<bool>((ref) {
  final currentTable = ref.watch(currentTableProvider);
  return currentTable.value != null;
});

// Provider for getting the current player
final currentPlayerProvider = Provider<PokerPlayer?>((ref) {
  final gameState = ref.watch(gameStateProvider);
  final webSocketService = ref.watch(webSocketServiceProvider);

  if (gameState.value == null) return null;

  final userId = webSocketService.connectionInfo.userID;
  if (userId == null) return null;

  return gameState.value!.getPlayerById(userId);
});

// Provider for checking if it's the current player's turn
final isMyTurnProvider = Provider<bool>((ref) {
  final gameState = ref.watch(gameStateProvider);
  final currentPlayer = ref.watch(currentPlayerProvider);

  if (gameState.value == null || currentPlayer == null) return false;

  return gameState.value!.activePlayerID == currentPlayer.id;
});

// Provider for available actions
final availableActionsProvider = Provider<List<PokerActionType>>((ref) {
  final gameStateNotifier = ref.watch(gameStateProvider.notifier);
  final isMyTurn = ref.watch(isMyTurnProvider);

  if (!isMyTurn) return [];

  final actions = <PokerActionType>[];

  for (final action in PokerActionType.values) {
    if (gameStateNotifier.canPerformAction(action)) {
      actions.add(action);
    }
  }

  return actions;
});
