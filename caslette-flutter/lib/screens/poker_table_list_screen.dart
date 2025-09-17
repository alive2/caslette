import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/poker/poker_models.dart' as poker;
import '../providers/poker_provider.dart';
import '../providers/auth_provider.dart';
import 'poker_game_screen.dart';

class PokerTableListScreen extends ConsumerStatefulWidget {
  const PokerTableListScreen({super.key});

  @override
  ConsumerState<PokerTableListScreen> createState() =>
      _PokerTableListScreenState();
}

class _PokerTableListScreenState extends ConsumerState<PokerTableListScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _initializePokerConnection();
    });
  }

  void _initializePokerConnection() async {
    final authState = ref.read(authProvider);
    final currentUser = ref.read(authProvider.notifier).currentUser;

    if (authState == AuthState.authenticated && currentUser != null) {
      final pokerNotifier = ref.read(pokerProvider.notifier);
      final pokerState = ref.read(pokerProvider);

      // Connect to poker WebSocket if not already connected
      if (!pokerState.isConnected) {
        await pokerNotifier.connect(currentUser.token);
      } else {
        // Already connected, just request table list
        pokerNotifier.requestTableList();
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final availableTables = ref.watch(availableTablesProvider);
    final isLoading = ref.watch(pokerLoadingProvider);
    final error = ref.watch(pokerErrorProvider);
    final isConnected = ref.watch(pokerConnectionProvider);

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
          child: Column(
            children: [
              // Header
              _buildHeader(isConnected),

              // Content
              Expanded(child: _buildContent(availableTables, isLoading, error)),
            ],
          ),
        ),
      ),
      floatingActionButton: _buildRefreshButton(),
    );
  }

  Widget _buildHeader(bool isConnected) {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Row(
        children: [
          IconButton(
            onPressed: () => Navigator.of(context).pop(),
            icon: const Icon(Icons.arrow_back, color: Colors.white),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  'Poker Tables',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 24,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  'Find and join a Texas Hold\'em table',
                  style: TextStyle(
                    color: Colors.white.withOpacity(0.7),
                    fontSize: 14,
                  ),
                ),
              ],
            ),
          ),
          // Connection status
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            decoration: BoxDecoration(
              color: isConnected ? Colors.green : Colors.red,
              borderRadius: BorderRadius.circular(12),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(
                  isConnected ? Icons.wifi : Icons.wifi_off,
                  color: Colors.white,
                  size: 14,
                ),
                const SizedBox(width: 4),
                Text(
                  isConnected ? 'Connected' : 'Offline',
                  style: const TextStyle(
                    color: Colors.white,
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildContent(
    List<poker.PokerTable> tables,
    bool isLoading,
    String? error,
  ) {
    if (error != null) {
      return _buildErrorState(error);
    }

    if (isLoading && tables.isEmpty) {
      return _buildLoadingState();
    }

    if (!ref.watch(pokerConnectionProvider)) {
      return _buildDisconnectedState();
    }

    if (tables.isEmpty) {
      return _buildEmptyState();
    }

    return _buildTableList(tables);
  }

  Widget _buildErrorState(String error) {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.error_outline, color: Colors.red, size: 64),
            const SizedBox(height: 16),
            const Text(
              'Failed to load tables',
              style: TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              error,
              style: TextStyle(
                color: Colors.white.withOpacity(0.7),
                fontSize: 14,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () {
                ref.read(pokerProvider.notifier).clearError();
                _initializePokerConnection();
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF059669),
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 12,
                ),
              ),
              child: const Text('Retry'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLoadingState() {
    return const Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          CircularProgressIndicator(
            valueColor: AlwaysStoppedAnimation<Color>(Color(0xFF059669)),
          ),
          SizedBox(height: 16),
          Text(
            'Loading poker tables...',
            style: TextStyle(color: Colors.white, fontSize: 16),
          ),
        ],
      ),
    );
  }

  Widget _buildDisconnectedState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(Icons.wifi_off, color: Colors.orange, size: 64),
            const SizedBox(height: 16),
            const Text(
              'Not Connected',
              style: TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'Unable to connect to poker server',
              style: TextStyle(
                color: Colors.white.withOpacity(0.7),
                fontSize: 14,
              ),
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: _initializePokerConnection,
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF059669),
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 12,
                ),
              ),
              child: const Text('Connect'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            const Icon(
              Icons.table_restaurant,
              color: Color(0xFF6B7280),
              size: 64,
            ),
            const SizedBox(height: 16),
            const Text(
              'No Tables Available',
              style: TextStyle(
                color: Colors.white,
                fontSize: 20,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 8),
            Text(
              'There are currently no poker tables available.\nCheck back later or refresh to try again.',
              style: TextStyle(
                color: Colors.white.withOpacity(0.7),
                fontSize: 14,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 24),
            ElevatedButton(
              onPressed: () =>
                  ref.read(pokerProvider.notifier).requestTableList(),
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFF059669),
                padding: const EdgeInsets.symmetric(
                  horizontal: 24,
                  vertical: 12,
                ),
              ),
              child: const Text('Refresh'),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTableList(List<poker.PokerTable> tables) {
    return RefreshIndicator(
      onRefresh: () async {
        ref.read(pokerProvider.notifier).requestTableList();
      },
      color: const Color(0xFF059669),
      backgroundColor: const Color(0xFF1F2937),
      child: ListView.builder(
        padding: const EdgeInsets.all(16),
        itemCount: tables.length,
        itemBuilder: (context, index) {
          final table = tables[index];
          return _buildTableCard(table);
        },
      ),
    );
  }

  Widget _buildTableCard(poker.PokerTable table) {
    final isActive = table.status == poker.TableStatus.active;
    final hasSpace = table.hasAvailableSeats;

    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF1F2937),
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: const Color(0xFF374151), width: 1),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          borderRadius: BorderRadius.circular(16),
          onTap: hasSpace ? () => _joinTable(table) : null,
          child: Padding(
            padding: const EdgeInsets.all(16),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            table.name,
                            style: const TextStyle(
                              color: Colors.white,
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(height: 4),
                          Text(
                            'Blinds: ${table.smallBlind}/${table.bigBlind}',
                            style: TextStyle(
                              color: Colors.white.withOpacity(0.7),
                              fontSize: 14,
                            ),
                          ),
                        ],
                      ),
                    ),
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: _getStatusColor(table.status).withOpacity(0.2),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        table.status.displayName,
                        style: TextStyle(
                          color: _getStatusColor(table.status),
                          fontSize: 12,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    _buildInfoChip(
                      Icons.people,
                      '${table.currentPlayers}/${table.maxPlayers}',
                      hasSpace
                          ? const Color(0xFF059669)
                          : const Color(0xFF9CA3AF),
                    ),
                    const SizedBox(width: 12),
                    _buildInfoChip(
                      Icons.monetization_on,
                      'Buy-in: ${table.minBuyIn}-${table.maxBuyIn}',
                      const Color(0xFFF59E0B),
                    ),
                    if (isActive) ...[
                      const SizedBox(width: 12),
                      _buildInfoChip(
                        Icons.timer,
                        table.bettingRound.displayName,
                        const Color(0xFF9333EA),
                      ),
                    ],
                  ],
                ),
                if (!hasSpace)
                  Container(
                    margin: const EdgeInsets.only(top: 12),
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: Colors.red.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(
                        color: Colors.red.withOpacity(0.3),
                        width: 1,
                      ),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.block, color: Colors.red, size: 16),
                        const SizedBox(width: 8),
                        Text(
                          'Table is full',
                          style: TextStyle(
                            color: Colors.red.shade300,
                            fontSize: 12,
                            fontWeight: FontWeight.w500,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildInfoChip(IconData icon, String text, Color color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: color),
          const SizedBox(width: 4),
          Text(
            text,
            style: TextStyle(
              color: color,
              fontSize: 11,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Color _getStatusColor(poker.TableStatus status) {
    switch (status) {
      case poker.TableStatus.waiting:
        return const Color(0xFFF59E0B);
      case poker.TableStatus.active:
        return const Color(0xFF059669);
      case poker.TableStatus.finished:
        return const Color(0xFF9CA3AF);
    }
  }

  Widget _buildRefreshButton() {
    return FloatingActionButton(
      onPressed: () => ref.read(pokerProvider.notifier).requestTableList(),
      backgroundColor: const Color(0xFF059669),
      child: const Icon(Icons.refresh, color: Colors.white),
    );
  }

  void _joinTable(poker.PokerTable table) {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (context) => PokerGameScreen(tableId: table.id),
      ),
    );
  }
}
