import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';
import '../providers/diamond_provider.dart';
import 'websocket_test_page.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _initializeData();
    });
  }

  void _initializeData() {
    final authNotifier = ref.read(authProvider.notifier);
    final currentUser = authNotifier.currentUser;

    if (currentUser != null) {
      ref
          .read(diamondProvider.notifier)
          .fetchBalanceFromAPI(currentUser.id, currentUser.token);
    }
  }

  @override
  Widget build(BuildContext context) {
    final diamondState = ref.watch(diamondProvider);
    final authNotifier = ref.read(authProvider.notifier);
    final currentUser = authNotifier.currentUser;

    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
            colors: [Color(0xFF1a3a3a), Color(0xFF0d2626)],
          ),
        ),
        child: SafeArea(
          child: LayoutBuilder(
            builder: (context, constraints) {
              final isLandscape = constraints.maxWidth > constraints.maxHeight;

              if (isLandscape) {
                return _buildLandscapeLayout(diamondState, currentUser);
              } else {
                return _buildPortraitLayout(diamondState, currentUser);
              }
            },
          ),
        ),
      ),
    );
  }

  Widget _buildPortraitLayout(DiamondState diamondState, User? currentUser) {
    return Column(
      children: [
        _buildHeader(diamondState, currentUser),
        Expanded(child: _buildMainContent()),
      ],
    );
  }

  Widget _buildLandscapeLayout(DiamondState diamondState, User? currentUser) {
    return Row(
      children: [
        Expanded(
          flex: 1,
          child: Column(
            children: [
              _buildHeader(diamondState, currentUser),
              Expanded(child: _buildMainContent()),
            ],
          ),
        ),
        Expanded(
          flex: 1,
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: _buildGameArea(),
          ),
        ),
      ],
    );
  }

  Widget _buildHeader(DiamondState diamondState, User? currentUser) {
    return Container(
      padding: const EdgeInsets.all(20),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'CASLETTE',
                  style: TextStyle(
                    fontSize: 32,
                    fontWeight: FontWeight.bold,
                    color: Colors.white.withOpacity(0.9),
                  ),
                ),
                if (currentUser != null)
                  Text(
                    'Welcome, ${currentUser.username}',
                    style: TextStyle(
                      fontSize: 16,
                      color: Colors.white.withOpacity(0.7),
                    ),
                  ),
              ],
            ),
          ),
          _buildBalanceCard(diamondState),
          const SizedBox(width: 16),
          IconButton(
            onPressed: () {
              Navigator.of(context).push(
                MaterialPageRoute(
                  builder: (context) => const WebSocketTestPage(),
                ),
              );
            },
            icon: Icon(
              Icons.developer_mode,
              color: Colors.white.withOpacity(0.8),
            ),
            tooltip: 'Websocket Test',
          ),
          IconButton(
            onPressed: () {
              ref.read(authProvider.notifier).logout();
            },
            icon: Icon(Icons.logout, color: Colors.white.withOpacity(0.8)),
            tooltip: 'Logout',
          ),
        ],
      ),
    );
  }

  Widget _buildBalanceCard(DiamondState diamondState) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.3),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.white.withOpacity(0.2), width: 1),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.diamond, color: Colors.amber, size: 24),
          const SizedBox(width: 8),
          diamondState.isLoading
              ? SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(
                    strokeWidth: 2,
                    valueColor: AlwaysStoppedAnimation<Color>(
                      Colors.white.withOpacity(0.8),
                    ),
                  ),
                )
              : Text(
                  '${diamondState.balance}',
                  style: TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.white.withOpacity(0.9),
                  ),
                ),
        ],
      ),
    );
  }

  Widget _buildMainContent() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildWelcome(),
          const SizedBox(height: 32),
          // Only show game area in portrait mode
          if (MediaQuery.of(context).orientation == Orientation.portrait)
            _buildGameArea(),
        ],
      ),
    );
  }

  Widget _buildWelcome() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Welcome to Caslette',
          style: TextStyle(
            fontSize: 28,
            fontWeight: FontWeight.bold,
            color: Colors.white.withOpacity(0.9),
          ),
        ),
        const SizedBox(height: 8),
        Text(
          'Your premium gaming destination',
          style: TextStyle(fontSize: 16, color: Colors.white.withOpacity(0.7)),
        ),
      ],
    );
  }

  Widget _buildGameArea() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          'Games',
          style: TextStyle(
            fontSize: 24,
            fontWeight: FontWeight.bold,
            color: Colors.white.withOpacity(0.9),
          ),
        ),
        const SizedBox(height: 16),
        Container(
          width: double.infinity,
          padding: const EdgeInsets.all(24),
          decoration: BoxDecoration(
            color: Colors.black.withOpacity(0.2),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.white.withOpacity(0.1), width: 1),
          ),
          child: Column(
            children: [
              Icon(Icons.games, size: 48, color: Colors.white.withOpacity(0.6)),
              const SizedBox(height: 16),
              Text(
                'Coming Soon',
                style: TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                  color: Colors.white.withOpacity(0.8),
                ),
              ),
              const SizedBox(height: 8),
              Text(
                'Exciting games are being prepared for you.',
                style: TextStyle(
                  fontSize: 16,
                  color: Colors.white.withOpacity(0.6),
                ),
                textAlign: TextAlign.center,
              ),
            ],
          ),
        ),
      ],
    );
  }
}
