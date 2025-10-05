import "package:flutter_riverpod/flutter_riverpod.dart";
import "../services/websocket/websocket_providers.dart";

class DiamondState {
  final int balance;
  final bool isLoading;
  final String? error;

  const DiamondState({
    required this.balance,
    this.isLoading = false,
    this.error,
  });

  DiamondState copyWith({int? balance, bool? isLoading, String? error}) {
    return DiamondState(
      balance: balance ?? this.balance,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

class DiamondNotifier extends StateNotifier<DiamondState> {
  final Ref _ref;

  DiamondNotifier(this._ref)
    : super(const DiamondState(balance: 0, isLoading: true));

  void updateBalance(int newBalance) {
    state = state.copyWith(balance: newBalance);
  }

  void addDiamonds(int amount) {
    final newBalance = state.balance + amount;
    updateBalance(newBalance);
  }

  void removeDiamonds(int amount) {
    final newBalance = state.balance - amount;
    if (newBalance >= 0) {
      updateBalance(newBalance);
    }
  }

  Future<void> fetchBalanceFromWebSocket(String userId) async {
    print(
      'DiamondProvider: fetchBalanceFromWebSocket called for userId: $userId',
    );
    state = state.copyWith(isLoading: true);

    try {
      final webSocketService = _ref.read(webSocketServiceProvider);
      print('DiamondProvider: Calling webSocketService.getUserBalance...');
      final response = await webSocketService.getUserBalance(userId);
      print(
        'DiamondProvider: Received response - success: ${response.success}, data: ${response.data}',
      );

      if (response.success && response.data != null) {
        final data = response.data as Map<String, dynamic>;
        if (data.containsKey('current_balance')) {
          final balance = data['current_balance'] as int;
          print('DiamondProvider: Setting balance to $balance');
          state = state.copyWith(
            balance: balance,
            isLoading: false,
            error: null,
          );
        } else {
          print(
            'DiamondProvider: Invalid balance data - missing current_balance key',
          );
          state = state.copyWith(
            isLoading: false,
            error: 'Invalid balance data received',
          );
        }
      } else {
        print('DiamondProvider: Request failed - error: ${response.error}');
        state = state.copyWith(
          isLoading: false,
          error: response.error ?? 'Failed to fetch balance via WebSocket',
        );
      }
    } catch (e) {
      print('DiamondProvider: Exception caught: $e');
      state = state.copyWith(
        isLoading: false,
        error: 'WebSocket error: ${e.toString()}',
      );
    }
  }

  // Keep the old method for backward compatibility, but mark it as deprecated
  @Deprecated('Use fetchBalanceFromWebSocket instead')
  Future<void> fetchBalanceFromAPI(String userId, String token) async {
    // For now, just call the WebSocket version
    await fetchBalanceFromWebSocket(userId);
  }
}

final diamondProvider = StateNotifierProvider<DiamondNotifier, DiamondState>((
  ref,
) {
  return DiamondNotifier(ref);
});
