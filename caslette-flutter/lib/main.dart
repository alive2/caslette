import "package:flutter/material.dart";
import "package:flutter_riverpod/flutter_riverpod.dart";
import "screens/home_screen.dart";
import "screens/login_screen.dart";
import "screens/poker_lobby_screen.dart";
import "screens/poker_table_screen.dart";
import "providers/auth_provider.dart";

void main() {
  runApp(const ProviderScope(child: MyApp()));
}

class MyApp extends ConsumerWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final authState = ref.watch(authProvider);

    return MaterialApp(
      title: "Caslette Casino",
      theme: ThemeData.dark(),
      home: _getHomeScreen(authState),
      routes: {
        '/poker': (context) => const PokerLobbyScreen(),
        '/poker/lobby': (context) => const PokerLobbyScreen(),
      },
      onGenerateRoute: (settings) {
        if (settings.name == '/poker/table') {
          return MaterialPageRoute(
            builder: (context) => const PokerTableScreen(),
            settings: settings,
          );
        }
        return null;
      },
    );
  }

  Widget _getHomeScreen(AuthState authState) {
    switch (authState) {
      case AuthState.authenticated:
        return const HomeScreen();
      case AuthState.loading:
        return const Scaffold(body: Center(child: CircularProgressIndicator()));
      case AuthState.unauthenticated:
        return const LoginScreen();
    }
  }
}
