import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../providers/poker_provider.dart';
import '../providers/auth_provider.dart';
import 'poker_table_list_screen.dart';

class PokerConnectionScreen extends ConsumerStatefulWidget {
  const PokerConnectionScreen({super.key});

  @override
  ConsumerState<PokerConnectionScreen> createState() =>
      _PokerConnectionScreenState();
}

class _PokerConnectionScreenState extends ConsumerState<PokerConnectionScreen>
    with SingleTickerProviderStateMixin {
  late AnimationController _animationController;
  late Animation<double> _pulseAnimation;
  String _connectionStatus = 'Initializing...';
  bool _isConnected = false;

  @override
  void initState() {
    super.initState();

    // Setup animation for the connection indicator
    _animationController = AnimationController(
      duration: const Duration(seconds: 2),
      vsync: this,
    )..repeat(reverse: true);

    _pulseAnimation = Tween<double>(begin: 0.8, end: 1.2).animate(
      CurvedAnimation(parent: _animationController, curve: Curves.easeInOut),
    );

    // Start connection process
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _startConnection();
    });
  }

  @override
  void dispose() {
    _animationController.dispose();
    super.dispose();
  }

  Future<void> _startConnection() async {
    final currentUser = ref.read(authProvider.notifier).currentUser;

    if (currentUser == null) {
      setState(() {
        _connectionStatus = 'Authentication required';
      });
      return;
    }

    setState(() {
      _connectionStatus = 'Connecting to poker server...';
    });

    try {
      final pokerNotifier = ref.read(pokerProvider.notifier);

      // Connect to WebSocket
      await pokerNotifier.connect(currentUser.token);

      // Wait a moment for connection to stabilize
      await Future.delayed(const Duration(milliseconds: 500));

      // Check if connection was successful
      final isConnected = ref.read(pokerConnectionProvider);

      if (isConnected) {
        setState(() {
          _connectionStatus = 'Connected! Loading poker tables...';
          _isConnected = true;
        });

        // Load initial tables via HTTP API
        await pokerNotifier.fetchTablesViaAPI(currentUser.token);

        // Wait a bit to show success message
        await Future.delayed(const Duration(milliseconds: 1000));

        // Navigate to poker tables screen
        if (mounted) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(
              builder: (context) => const PokerTableListScreen(),
            ),
          );
        }
      } else {
        setState(() {
          _connectionStatus = 'Connection failed. Tap to retry.';
        });
      }
    } catch (e) {
      setState(() {
        _connectionStatus = 'Connection error: ${e.toString()}\nTap to retry.';
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0F172A),
      body: Container(
        decoration: const BoxDecoration(
          gradient: RadialGradient(
            center: Alignment.center,
            radius: 1.2,
            colors: [
              Color(0xFF064E3B), // Dark emerald
              Color(0xFF0F172A), // Slate-900
            ],
          ),
        ),
        child: SafeArea(
          child: Center(
            child: Padding(
              padding: const EdgeInsets.all(32.0),
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  // Logo/Title
                  const Text(
                    'Caslette Poker',
                    style: TextStyle(
                      fontSize: 32,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  const SizedBox(height: 60),

                  // Connection indicator
                  AnimatedBuilder(
                    animation: _pulseAnimation,
                    builder: (context, child) {
                      return Transform.scale(
                        scale: _isConnected ? 1.0 : _pulseAnimation.value,
                        child: Container(
                          width: 120,
                          height: 120,
                          decoration: BoxDecoration(
                            shape: BoxShape.circle,
                            color: _isConnected
                                ? const Color(0xFF059669)
                                : const Color(0xFF374151),
                            border: Border.all(
                              color: _isConnected
                                  ? const Color(0xFF10B981)
                                  : const Color(0xFF6B7280),
                              width: 3,
                            ),
                            boxShadow: [
                              BoxShadow(
                                color:
                                    (_isConnected
                                            ? const Color(0xFF059669)
                                            : const Color(0xFF374151))
                                        .withOpacity(0.3),
                                blurRadius: 20,
                                spreadRadius: 5,
                              ),
                            ],
                          ),
                          child: Icon(
                            _isConnected ? Icons.check : Icons.wifi,
                            size: 50,
                            color: Colors.white,
                          ),
                        ),
                      );
                    },
                  ),

                  const SizedBox(height: 40),

                  // Status text
                  Text(
                    _connectionStatus,
                    textAlign: TextAlign.center,
                    style: const TextStyle(
                      fontSize: 18,
                      color: Colors.white70,
                      height: 1.4,
                    ),
                  ),

                  const SizedBox(height: 40),

                  // Retry button (only show if connection failed)
                  if (_connectionStatus.contains('retry') ||
                      _connectionStatus.contains('error'))
                    ElevatedButton(
                      onPressed: () {
                        setState(() {
                          _isConnected = false;
                        });
                        _startConnection();
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFF059669),
                        foregroundColor: Colors.white,
                        padding: const EdgeInsets.symmetric(
                          horizontal: 32,
                          vertical: 16,
                        ),
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                      child: const Text(
                        'Retry Connection',
                        style: TextStyle(fontSize: 16),
                      ),
                    ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
