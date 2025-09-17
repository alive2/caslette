import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/auth_provider.dart';
import '../providers/diamond_provider.dart';
import '../services/websocket_service.dart';
import 'poker_connection_screen.dart';

class HomeScreen extends ConsumerStatefulWidget {
  const HomeScreen({super.key});

  @override
  ConsumerState<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends ConsumerState<HomeScreen> {
  @override
  void initState() {
    super.initState();
    // Initialize WebSocket and fetch balance when screen loads
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _initializeData();
    });
  }

  void _initializeData() {
    final currentUser = ref.read(authProvider.notifier).currentUser;
    if (currentUser != null) {
      // Connect WebSocket
      ref.read(diamondProvider.notifier).connectWebSocket(currentUser.token);
      // Fetch initial balance
      ref
          .read(diamondProvider.notifier)
          .fetchBalanceFromAPI(currentUser.id, currentUser.token);
    }
  }

  @override
  Widget build(BuildContext context) {
    final diamondState = ref.watch(diamondProvider);
    final wsService = ref.watch(webSocketServiceProvider);
    final authNotifier = ref.read(authProvider.notifier);
    final currentUser = authNotifier.currentUser;

    return Scaffold(
      backgroundColor: const Color(0xFF0F172A), // Slate-900
      body: Container(
        decoration: const BoxDecoration(
          gradient: RadialGradient(
            center: Alignment.center,
            radius: 1.2,
            colors: [
              Color(0xFF4C1D95), // Purple-800
              Color(0xFF0F172A), // Slate-900
            ],
          ),
        ),
        child: SafeArea(
          child: OrientationBuilder(
            builder: (context, orientation) {
              if (orientation == Orientation.landscape) {
                return _buildLandscapeLayout(
                  diamondState,
                  wsService,
                  currentUser,
                );
              } else {
                return _buildPortraitLayout(
                  diamondState,
                  wsService,
                  currentUser,
                );
              }
            },
          ),
        ),
      ),
    );
  }

  Widget _buildPortraitLayout(
    DiamondState diamondState,
    WebSocketService wsService,
    User? currentUser,
  ) {
    return Column(
      children: [
        _buildHeader(wsService, currentUser),
        _buildBalanceCard(diamondState),
        const SizedBox(height: 32),
        Expanded(child: _buildPokerSection()),
        const SizedBox(height: 20),
      ],
    );
  }

  Widget _buildLandscapeLayout(
    DiamondState diamondState,
    WebSocketService wsService,
    User? currentUser,
  ) {
    return Row(
      children: [
        // Left side - Header and Balance
        Expanded(
          flex: 1,
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: Column(
              children: [
                _buildHeader(wsService, currentUser, isLandscape: true),
                const SizedBox(height: 20),
                _buildBalanceCard(diamondState, isLandscape: true),
                const Spacer(),
              ],
            ),
          ),
        ),
        // Right side - Poker Game
        Expanded(
          flex: 1,
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: _buildPokerSection(isLandscape: true),
          ),
        ),
      ],
    );
  }

  Widget _buildHeader(
    WebSocketService wsService,
    User? currentUser, {
    bool isLandscape = false,
  }) {
    return Container(
      padding: EdgeInsets.all(isLandscape ? 16 : 20),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Row(
              children: [
                Icon(
                  Icons.casino,
                  color: const Color(0xFF9333EA),
                  size: isLandscape ? 24 : 32,
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'CASLETTE',
                        style: TextStyle(
                          fontSize: isLandscape ? 18 : 24,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                          letterSpacing: 1,
                        ),
                      ),
                      if (currentUser != null)
                        Text(
                          'Welcome, ${currentUser.username}',
                          style: TextStyle(
                            fontSize: isLandscape ? 12 : 14,
                            color: const Color(0xFF9CA3AF),
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                    ],
                  ),
                ),
              ],
            ),
          ),
          Row(
            children: [
              // WebSocket status indicator
              Container(
                padding: EdgeInsets.symmetric(
                  horizontal: isLandscape ? 6 : 8,
                  vertical: 4,
                ),
                decoration: BoxDecoration(
                  color: wsService.isConnected ? Colors.green : Colors.red,
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      wsService.isConnected ? Icons.wifi : Icons.wifi_off,
                      color: Colors.white,
                      size: isLandscape ? 14 : 16,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      wsService.isConnected ? 'Connected' : 'Offline',
                      style: TextStyle(
                        color: Colors.white,
                        fontSize: isLandscape ? 10 : 12,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 12),
              IconButton(
                onPressed: () {
                  ref.read(authProvider.notifier).logout();
                },
                icon: Icon(
                  Icons.logout,
                  color: const Color(0xFF9CA3AF),
                  size: isLandscape ? 20 : 24,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBalanceCard(
    DiamondState diamondState, {
    bool isLandscape = false,
  }) {
    return Container(
      margin: EdgeInsets.symmetric(horizontal: isLandscape ? 0 : 20),
      padding: EdgeInsets.symmetric(
        horizontal: isLandscape ? 20 : 24,
        vertical: isLandscape ? 20 : 24,
      ),
      constraints: isLandscape ? const BoxConstraints(minHeight: 80) : null,
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [
            Color(0xFF9333EA), // Purple-600
            Color(0xFFF59E0B), // Amber-500
          ],
        ),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: const Color(0xFF9333EA).withOpacity(0.3),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: isLandscape
          ? Row(
              children: [
                // Left side - Balance info
                Expanded(
                  flex: 2,
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Text(
                        'Cosmic Balance',
                        style: TextStyle(fontSize: 16, color: Colors.white70),
                      ),
                      const SizedBox(height: 8),
                      Row(
                        children: [
                          const Icon(
                            Icons.diamond,
                            color: Colors.white,
                            size: 28,
                          ),
                          const SizedBox(width: 8),
                          Flexible(
                            child: Text(
                              '${diamondState.balance}',
                              style: const TextStyle(
                                fontSize: 28,
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                // Right side - Decorative icon
                Expanded(
                  flex: 1,
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(
                        Icons.auto_awesome,
                        color: Colors.white,
                        size: 40,
                      ),
                      const SizedBox(height: 4),
                      Text(
                        'DIAMONDS',
                        style: TextStyle(
                          fontSize: 10,
                          color: Colors.white.withOpacity(0.7),
                          fontWeight: FontWeight.bold,
                          letterSpacing: 1,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            )
          : Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      'Cosmic Balance',
                      style: TextStyle(fontSize: 16, color: Colors.white70),
                    ),
                    const SizedBox(height: 8),
                    Row(
                      children: [
                        const Icon(
                          Icons.diamond,
                          color: Colors.white,
                          size: 24,
                        ),
                        const SizedBox(width: 8),
                        Text(
                          '${diamondState.balance}',
                          style: const TextStyle(
                            fontSize: 32,
                            fontWeight: FontWeight.bold,
                            color: Colors.white,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
                Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    const Icon(
                      Icons.auto_awesome,
                      color: Colors.white,
                      size: 48,
                    ),
                    const SizedBox(height: 4),
                    Text(
                      'DIAMONDS',
                      style: TextStyle(
                        fontSize: 10,
                        color: Colors.white.withOpacity(0.7),
                        fontWeight: FontWeight.bold,
                        letterSpacing: 1,
                      ),
                    ),
                  ],
                ),
              ],
            ),
    );
  }

  Widget _buildPokerSection({bool isLandscape = false}) {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: isLandscape ? 0 : 20),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Texas Hold\'em Poker',
            style: TextStyle(
              fontSize: isLandscape ? 20 : 24,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
          SizedBox(height: isLandscape ? 12 : 16),
          Expanded(
            child: Center(
              child: SizedBox(
                width: isLandscape ? 160 : 200,
                height: isLandscape ? 160 : 200,
                child: _buildPokerGameCard(
                  'Texas Hold\'em',
                  'Most popular poker variant',
                  Icons.style,
                  const Color(0xFF059669), // Emerald-600
                  () {
                    Navigator.of(context).push(
                      MaterialPageRoute(
                        builder: (context) => const PokerConnectionScreen(),
                      ),
                    );
                  },
                  isLandscape: isLandscape,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPokerGameCard(
    String title,
    String description,
    IconData icon,
    Color color,
    VoidCallback onTap, {
    bool isLandscape = false,
  }) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        decoration: BoxDecoration(
          color: const Color(0xFF1F2937), // Gray-800
          borderRadius: BorderRadius.circular(16),
          border: Border.all(
            color: const Color(0xFF374151), // Gray-700
            width: 1,
          ),
          boxShadow: [
            BoxShadow(
              color: color.withOpacity(0.2),
              blurRadius: 10,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Padding(
          padding: EdgeInsets.all(isLandscape ? 12 : 16),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                padding: EdgeInsets.all(isLandscape ? 8 : 12),
                decoration: BoxDecoration(
                  color: color.withOpacity(0.2),
                  shape: BoxShape.circle,
                ),
                child: Icon(icon, size: isLandscape ? 24 : 28, color: color),
              ),
              SizedBox(height: isLandscape ? 8 : 12),
              Text(
                title,
                style: TextStyle(
                  fontSize: isLandscape ? 14 : 16,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 4),
              Text(
                description,
                style: TextStyle(
                  fontSize: isLandscape ? 10 : 11,
                  color: Colors.white.withOpacity(0.7),
                ),
                textAlign: TextAlign.center,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ],
          ),
        ),
      ),
    );
  }
}
