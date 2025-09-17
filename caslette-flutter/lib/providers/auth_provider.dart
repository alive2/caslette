import "package:flutter_riverpod/flutter_riverpod.dart";
import "package:shared_preferences/shared_preferences.dart";
import "../services/api_service.dart";
import "diamond_provider.dart";

enum AuthState { unauthenticated, authenticated, loading }

class User {
  final String id;
  final String username;
  final String token;

  User({required this.id, required this.username, required this.token});

  factory User.fromJson(Map<String, dynamic> json) {
    return User(
      id: json['id'] ?? '',
      username: json['username'] ?? '',
      token: json['token'] ?? '',
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  final Ref _ref;
  User? _currentUser;

  AuthNotifier(this._ref) : super(AuthState.loading) {
    _checkAuthStatus();
  }

  User? get currentUser => _currentUser;

  Future<void> _checkAuthStatus() async {
    try {
      final prefs = await SharedPreferences.getInstance();
      final token = prefs.getString('auth_token');
      final username = prefs.getString('username');
      final userId = prefs.getString('user_id');

      if (token != null && username != null && userId != null) {
        _currentUser = User(id: userId, username: username, token: token);
        state = AuthState.authenticated;
      } else {
        state = AuthState.unauthenticated;
      }
    } catch (e) {
      state = AuthState.unauthenticated;
    }
  }

  Future<bool> login(String username, String password) async {
    state = AuthState.loading;

    try {
      // Always use real API authentication
      final apiService = _ref.read(apiServiceProvider);
      final result = await apiService.login(username, password);

      if (result != null && result.containsKey('token')) {
        final user = User(
          id: result['user']['id'].toString(),
          username: username,
          token: result['token'],
        );

        _currentUser = user;

        // Store auth data
        final prefs = await SharedPreferences.getInstance();
        await prefs.setString('auth_token', user.token);
        await prefs.setString('username', user.username);
        await prefs.setString('user_id', user.id);

        state = AuthState.authenticated;
        return true;
      } else {
        state = AuthState.unauthenticated;
        return false;
      }
    } catch (e) {
      // Log the error for debugging
      print('Login error: $e');
      state = AuthState.unauthenticated;
      return false;
    }
  }

  Future<void> logout() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('auth_token');
    await prefs.remove('username');
    await prefs.remove('user_id');

    _currentUser = null;
    state = AuthState.unauthenticated;
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier(ref);
});
