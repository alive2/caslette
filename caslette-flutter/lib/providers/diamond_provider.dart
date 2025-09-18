import "package:flutter_riverpod/flutter_riverpod.dart";
import "../services/api_service.dart";

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

  Future<void> fetchBalanceFromAPI(String userId, String token) async {
    state = state.copyWith(isLoading: true);

    try {
      final apiService = _ref.read(apiServiceProvider);
      final result = await apiService.getBalance(userId, token);

      if (result != null && result.containsKey('current_balance')) {
        state = state.copyWith(
          balance: result['current_balance'] as int,
          isLoading: false,
        );
      } else {
        state = state.copyWith(
          isLoading: false,
          error: 'Failed to fetch balance',
        );
      }
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }
}

final diamondProvider = StateNotifierProvider<DiamondNotifier, DiamondState>((
  ref,
) {
  return DiamondNotifier(ref);
});
