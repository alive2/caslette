import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/poker/poker_models.dart' as poker;
import '../providers/poker_provider.dart';
import '../providers/auth_provider.dart';
import '../widgets/poker/table_widgets.dart';

class PokerGameScreen extends ConsumerStatefulWidget {
  final String tableId;

  const PokerGameScreen({super.key, required this.tableId});

  @override
  ConsumerState<PokerGameScreen> createState() => _PokerGameScreenState();
}

class _PokerGameScreenState extends ConsumerState<PokerGameScreen> {
  bool _hasJoinedTable = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _initializePokerGame();
    });
  }

  void _initializePokerGame() {
    final authState = ref.read(authProvider);
    final currentUser = ref.read(authProvider.notifier).currentUser;

    if (authState == AuthState.authenticated && currentUser != null) {
      final pokerNotifier = ref.read(pokerProvider.notifier);

      try {
        // Find the table from available tables
        final availableTables = ref.read(availableTablesProvider);
        final table = availableTables.firstWhere(
          (t) => t.id == widget.tableId,
          orElse: () => throw Exception('Table not found'),
        );

        print('DEBUG: Found table: ${table.name} (ID: ${table.id})');

        // Set this table as the current table
        pokerNotifier.setCurrentTable(table, widget.tableId);

        // Request game state for this table
        pokerNotifier.requestGameState();
      } catch (e) {
        print('DEBUG: Error finding table: $e');
        // Show error and navigate back
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Table not found: ${widget.tableId}'),
              backgroundColor: Colors.red,
            ),
          );
          Navigator.of(context).pop();
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final pokerState = ref.watch(pokerProvider);
    final currentUser = ref.read(authProvider.notifier).currentUser;
    final table = pokerState.currentTable;

    return Scaffold(
      backgroundColor: Colors.black,
      body: _buildGameContent(pokerState, currentUser?.id, table),
    );
  }

  Widget _buildGameContent(
    PokerState pokerState,
    String? currentUserId,
    poker.PokerTable? table,
  ) {
    if (pokerState.isLoading) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            CircularProgressIndicator(
              valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF059669)),
            ),
            SizedBox(height: 16),
            Text(
              'Loading poker table...',
              style: TextStyle(color: Colors.white, fontSize: 16),
            ),
          ],
        ),
      );
    }

    if (pokerState.error != null) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, color: Colors.red, size: 48),
            const SizedBox(height: 16),
            Text(
              'Error: ${pokerState.error}',
              style: const TextStyle(color: Colors.red, fontSize: 16),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: () {
                ref.read(pokerProvider.notifier).clearError();
                _initializePokerGame();
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF059669),
              ),
              child: const Text('Retry'),
            ),
          ],
        ),
      );
    }

    if (!pokerState.isConnected) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.wifi_off, color: Colors.orange, size: 48),
            const SizedBox(height: 16),
            const Text(
              'Connecting to poker server...',
              style: TextStyle(color: Colors.white, fontSize: 16),
            ),
            const SizedBox(height: 16),
            ElevatedButton(
              onPressed: _initializePokerGame,
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF059669),
              ),
              child: const Text('Reconnect'),
            ),
          ],
        ),
      );
    }

    if (table == null) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.table_restaurant, color: Colors.grey, size: 48),
            SizedBox(height: 16),
            Text(
              'Table not found',
              style: TextStyle(color: Colors.white, fontSize: 16),
            ),
          ],
        ),
      );
    }

    return PokerTableLayout(
      table: table,
      currentUserId: currentUserId,
      onJoinSeat: (seatNumber) => _showJoinSeatDialog(seatNumber, table),
      onPlayerAction: (action, {amount}) =>
          _handlePlayerAction(action, amount: amount),
    );
  }

  void _showJoinSeatDialog(int seatNumber, poker.PokerTable table) {
    final buyInController = TextEditingController(
      text: table.minBuyIn.toString(),
    );

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: const Color(0xFF1F2937),
        title: const Text(
          'Join Poker Table',
          style: TextStyle(color: Colors.white),
        ),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(
              'Seat $seatNumber',
              style: const TextStyle(color: Colors.white, fontSize: 16),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: buyInController,
              keyboardType: TextInputType.number,
              style: const TextStyle(color: Colors.white),
              decoration: InputDecoration(
                labelText: 'Buy-in Amount',
                labelStyle: const TextStyle(color: Color(0xFF9CA3AF)),
                hintText: 'Min: ${table.minBuyIn}, Max: ${table.maxBuyIn}',
                hintStyle: const TextStyle(color: Color(0xFF6B7280)),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: Color(0xFF374151)),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: Color(0xFF374151)),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: const BorderSide(color: Color(0xFF059669)),
                ),
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: const Text(
              'Cancel',
              style: TextStyle(color: Color(0xFF9CA3AF)),
            ),
          ),
          ElevatedButton(
            onPressed: () {
              final buyInAmount =
                  int.tryParse(buyInController.text) ?? table.minBuyIn;
              if (buyInAmount >= table.minBuyIn &&
                  buyInAmount <= table.maxBuyIn) {
                ref
                    .read(pokerProvider.notifier)
                    .joinTable(table.id, seatNumber, buyInAmount);
                setState(() {
                  _hasJoinedTable = true;
                });
                Navigator.of(context).pop();
              } else {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(
                    content: Text(
                      'Buy-in must be between ${table.minBuyIn} and ${table.maxBuyIn}',
                    ),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF059669),
            ),
            child: const Text('Join'),
          ),
        ],
      ),
    );
  }

  void _handlePlayerAction(poker.PlayerAction action, {int? amount}) {
    ref.read(pokerProvider.notifier).performAction(action, amount: amount);
  }

  @override
  void dispose() {
    // Leave table when disposing the screen if we joined it
    if (_hasJoinedTable) {
      ref.read(pokerProvider.notifier).leaveTable();
    }
    super.dispose();
  }
}
