import "package:flutter_riverpod/flutter_riverpod.dart";
import "package:shared_preferences/shared_preferences.dart";
import "../services/api_service.dart";
import "socket_provider.dart";
import "../services/websocket/websocket_providers.dart";

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
        final user = User(id: userId, username: username, token: token);
        _currentUser = user;
        state = AuthState.authenticated;

        // Connect to WebSocket for existing session
        _connectToSocket(user);
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

        // Connect to WebSocket after successful authentication
        _connectToSocket(user);

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

  // Connect to WebSocket and authenticate
  Future<void> _connectToSocket(User user) async {
    try {
      final webSocketController = _ref.read(
        webSocketControllerProvider.notifier,
      );

      // Re-enable WebSocket connections in case user was previously logged out
      final webSocketService = _ref.read(webSocketServiceProvider);
      webSocketService.enableReconnection();

      // Connect to WebSocket server
      await webSocketController.connect();

      // Authenticate with WebSocket
      final authenticated = await webSocketController.authenticate(user.token);

      if (authenticated) {
        print('WebSocket authentication successful');
      } else {
        print('WebSocket authentication failed');
      }
    } catch (e) {
      print('WebSocket connection/authentication error: $e');
      // Don't fail the login if WebSocket fails
    }
  }

  Future<void> logout() async {
    // Disconnect from WebSocket first
    try {
      final webSocketController = _ref.read(
        webSocketControllerProvider.notifier,
      );
      await webSocketController.disconnect();
    } catch (e) {
      print('WebSocket disconnect error: $e');
    }

    // Clear stored auth data
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
