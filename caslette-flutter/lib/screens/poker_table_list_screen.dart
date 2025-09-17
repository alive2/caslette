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
    print('DEBUG: PokerTableListScreen initState called');
    WidgetsBinding.instance.addPostFrameCallback((_) {
      print('DEBUG: initState postFrameCallback triggered');
      _loadInitialTables();
    });
  }

  void _loadInitialTables() async {
    print('DEBUG: _loadInitialTables called');

    final currentUser = ref.read(authProvider.notifier).currentUser;

    if (currentUser != null) {
      print('DEBUG: Loading tables for user: ${currentUser.username}');
      try {
        final pokerNotifier = ref.read(pokerProvider.notifier);
        await pokerNotifier.fetchTablesViaAPI(currentUser.token);
        print('DEBUG: Initial table loading completed');
      } catch (e) {
        print('DEBUG: Error loading initial tables: $e');
      }
    } else {
      print('DEBUG: No current user available for table loading');
    }
  }

  @override
  Widget build(BuildContext context) {
    final availableTables = ref.watch(availableTablesProvider);
    final isLoading = ref.watch(pokerLoadingProvider);
    final error = ref.watch(pokerErrorProvider);
    final isConnected = ref.watch(pokerConnectionProvider);

    print(
      'DEBUG: UI Build - isLoading: $isLoading, availableTables: ${availableTables.length}, isConnected: $isConnected',
    );

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
      floatingActionButton: Column(
        mainAxisAlignment: MainAxisAlignment.end,
        children: [
          FloatingActionButton(
            heroTag: "create_table",
            onPressed: () {
              if (isConnected) {
                _showCreateTableDialog();
              } else {
                ScaffoldMessenger.of(context).showSnackBar(
                  const SnackBar(
                    content: Text('Not connected to poker server'),
                    backgroundColor: Colors.red,
                  ),
                );
              }
            },
            backgroundColor: isConnected
                ? const Color(0xFF059669)
                : Colors.grey,
            child: const Icon(Icons.add, color: Colors.white),
          ),
          const SizedBox(height: 16),
          _buildRefreshButton(),
        ],
      ),
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
                _loadInitialTables();
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
              onPressed: _loadInitialTables,
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
              onPressed: () async {
                final currentUser = ref.read(authProvider.notifier).currentUser;
                if (currentUser != null) {
                  await ref
                      .read(pokerProvider.notifier)
                      .fetchTablesViaAPI(currentUser.token);
                }
              },
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
        final currentUser = ref.read(authProvider.notifier).currentUser;
        if (currentUser != null) {
          await ref
              .read(pokerProvider.notifier)
              .fetchTablesViaAPI(currentUser.token);
        }
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
      onPressed: () async {
        final currentUser = ref.read(authProvider.notifier).currentUser;
        if (currentUser != null) {
          await ref
              .read(pokerProvider.notifier)
              .fetchTablesViaAPI(currentUser.token);
        }
      },
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

  void _showCreateTableDialog() {
    // Check connection state before showing dialog
    final pokerState = ref.read(pokerProvider);
    final webSocketService = ref.read(pokerWebSocketServiceProvider);

    if (!pokerState.isConnected && !webSocketService.isConnected) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Not connected to poker server. Please wait...'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }
    final _nameController = TextEditingController();
    final _minBuyInController = TextEditingController(text: '100');
    final _maxBuyInController = TextEditingController(text: '1000');
    final _smallBlindController = TextEditingController(text: '5');
    final _bigBlindController = TextEditingController(text: '10');
    final _maxPlayersController = TextEditingController(text: '9');

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return Dialog.fullscreen(
          backgroundColor: Colors.transparent,
          child: Container(
            decoration: const BoxDecoration(
              gradient: LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  Color(0xFF1F2937),
                  Color(0xFF111827),
                  Color(0xFF0F172A),
                ],
              ),
            ),
            child: Column(
              children: [
                // Header
                Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 20,
                  ),
                  decoration: const BoxDecoration(
                    gradient: LinearGradient(
                      colors: [Color(0xFF059669), Color(0xFF047857)],
                    ),
                  ),
                  child: SafeArea(
                    bottom: false,
                    child: Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: const Icon(
                            Icons.casino,
                            color: Colors.white,
                            size: 24,
                          ),
                        ),
                        const SizedBox(width: 16),
                        const Expanded(
                          child: Text(
                            'Create New Table',
                            style: TextStyle(
                              color: Colors.white,
                              fontSize: 24,
                              fontWeight: FontWeight.bold,
                              letterSpacing: 0.5,
                            ),
                          ),
                        ),
                        IconButton(
                          onPressed: () => Navigator.of(context).pop(),
                          icon: const Icon(
                            Icons.close,
                            color: Colors.white,
                            size: 28,
                          ),
                          style: IconButton.styleFrom(
                            backgroundColor: Colors.white.withOpacity(0.1),
                            padding: const EdgeInsets.all(12),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                // Content
                Expanded(
                  child: SingleChildScrollView(
                    padding: const EdgeInsets.all(24),
                    child: Center(
                      child: ConstrainedBox(
                        constraints: const BoxConstraints(maxWidth: 1200),
                        child: LayoutBuilder(
                          builder: (context, constraints) {
                            final isLandscape = constraints.maxWidth > 700;

                            if (isLandscape) {
                              // Landscape layout - optimized two columns
                              return Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  // Left Column - Table Basic Info
                                  Expanded(
                                    flex: 5,
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const Text(
                                          'Table Setup',
                                          style: TextStyle(
                                            color: Color(0xFF059669),
                                            fontSize: 20,
                                            fontWeight: FontWeight.bold,
                                            letterSpacing: 0.5,
                                          ),
                                        ),
                                        const SizedBox(height: 20),

                                        // Table Name - Full width
                                        _buildFullscreenTextField(
                                          controller: _nameController,
                                          label: 'Table Name',
                                          icon: Icons.table_restaurant,
                                          hint: 'High Stakes Vegas',
                                        ),
                                        const SizedBox(height: 24),

                                        // Max Players - Full width
                                        _buildFullscreenTextField(
                                          controller: _maxPlayersController,
                                          label: 'Maximum Players (2-9)',
                                          icon: Icons.people,
                                          keyboardType: TextInputType.number,
                                          hint: '6',
                                        ),
                                        const SizedBox(height: 32),

                                        // Buy-in Section
                                        const Text(
                                          'Buy-in Limits',
                                          style: TextStyle(
                                            color: Color(0xFF059669),
                                            fontSize: 16,
                                            fontWeight: FontWeight.w600,
                                            letterSpacing: 0.3,
                                          ),
                                        ),
                                        const SizedBox(height: 16),

                                        // Buy-in Row - Side by side
                                        Row(
                                          children: [
                                            Expanded(
                                              child: _buildFullscreenTextField(
                                                controller: _minBuyInController,
                                                label: 'Minimum',
                                                icon: Icons.attach_money,
                                                keyboardType:
                                                    TextInputType.number,
                                                hint: '100',
                                              ),
                                            ),
                                            const SizedBox(width: 16),
                                            Expanded(
                                              child: _buildFullscreenTextField(
                                                controller: _maxBuyInController,
                                                label: 'Maximum',
                                                icon: Icons.money_rounded,
                                                keyboardType:
                                                    TextInputType.number,
                                                hint: '1000',
                                              ),
                                            ),
                                          ],
                                        ),
                                      ],
                                    ),
                                  ),

                                  const SizedBox(width: 24),

                                  // Right Column - Blinds & Game Settings
                                  Expanded(
                                    flex: 5,
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const Text(
                                          'Game Configuration',
                                          style: TextStyle(
                                            color: Color(0xFF059669),
                                            fontSize: 20,
                                            fontWeight: FontWeight.bold,
                                            letterSpacing: 0.5,
                                          ),
                                        ),
                                        const SizedBox(height: 20),

                                        // Blinds Section
                                        const Text(
                                          'Blind Structure',
                                          style: TextStyle(
                                            color: Color(0xFF059669),
                                            fontSize: 16,
                                            fontWeight: FontWeight.w600,
                                            letterSpacing: 0.3,
                                          ),
                                        ),
                                        const SizedBox(height: 16),

                                        // Blinds Row - Side by side
                                        Row(
                                          children: [
                                            Expanded(
                                              child: _buildFullscreenTextField(
                                                controller:
                                                    _smallBlindController,
                                                label: 'Small Blind',
                                                icon: Icons.casino_outlined,
                                                keyboardType:
                                                    TextInputType.number,
                                                hint: '5',
                                              ),
                                            ),
                                            const SizedBox(width: 16),
                                            Expanded(
                                              child: _buildFullscreenTextField(
                                                controller: _bigBlindController,
                                                label: 'Big Blind',
                                                icon: Icons.casino,
                                                keyboardType:
                                                    TextInputType.number,
                                                hint: '10',
                                              ),
                                            ),
                                          ],
                                        ),
                                        const SizedBox(height: 32),

                                        // Compact Game Info Panel
                                        Container(
                                          width: double.infinity,
                                          padding: const EdgeInsets.all(16),
                                          decoration: BoxDecoration(
                                            gradient: LinearGradient(
                                              begin: Alignment.topLeft,
                                              end: Alignment.bottomRight,
                                              colors: [
                                                const Color(
                                                  0xFF059669,
                                                ).withOpacity(0.12),
                                                const Color(
                                                  0xFF059669,
                                                ).withOpacity(0.04),
                                              ],
                                            ),
                                            borderRadius: BorderRadius.circular(
                                              12,
                                            ),
                                            border: Border.all(
                                              color: const Color(
                                                0xFF059669,
                                              ).withOpacity(0.3),
                                              width: 1.2,
                                            ),
                                          ),
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              Row(
                                                children: [
                                                  Container(
                                                    padding:
                                                        const EdgeInsets.all(6),
                                                    decoration: BoxDecoration(
                                                      color: const Color(
                                                        0xFF059669,
                                                      ).withOpacity(0.2),
                                                      borderRadius:
                                                          BorderRadius.circular(
                                                            6,
                                                          ),
                                                    ),
                                                    child: const Icon(
                                                      Icons.settings,
                                                      color: Color(0xFF059669),
                                                      size: 16,
                                                    ),
                                                  ),
                                                  const SizedBox(width: 8),
                                                  const Text(
                                                    'Default Settings',
                                                    style: TextStyle(
                                                      color: Color(0xFF059669),
                                                      fontSize: 14,
                                                      fontWeight:
                                                          FontWeight.bold,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                              const SizedBox(height: 12),
                                              _buildCompactInfoGrid([
                                                ['Game Type', 'Texas Hold\'em'],
                                                ['Rake', '5% (max \$50)'],
                                                ['Visibility', 'Public'],
                                                ['Auto-Start', '2+ players'],
                                                ['Action Timer', '30 seconds'],
                                                ['Deck', 'Standard 52'],
                                              ]),
                                            ],
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              );
                            } else {
                              // Portrait layout - single column
                              return Column(
                                children: [
                                  // Table Name
                                  _buildFullscreenTextField(
                                    controller: _nameController,
                                    label: 'Table Name',
                                    icon: Icons.table_restaurant,
                                    hint: 'High Stakes Vegas',
                                  ),
                                  const SizedBox(height: 24),

                                  // Buy-in Row
                                  Row(
                                    children: [
                                      Expanded(
                                        child: _buildFullscreenTextField(
                                          controller: _minBuyInController,
                                          label: 'Minimum Buy-in',
                                          icon: Icons.attach_money,
                                          keyboardType: TextInputType.number,
                                          hint: '100',
                                        ),
                                      ),
                                      const SizedBox(width: 20),
                                      Expanded(
                                        child: _buildFullscreenTextField(
                                          controller: _maxBuyInController,
                                          label: 'Maximum Buy-in',
                                          icon: Icons.money_rounded,
                                          keyboardType: TextInputType.number,
                                          hint: '1000',
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 24),

                                  // Blinds Row
                                  Row(
                                    children: [
                                      Expanded(
                                        child: _buildFullscreenTextField(
                                          controller: _smallBlindController,
                                          label: 'Small Blind',
                                          icon: Icons.casino_outlined,
                                          keyboardType: TextInputType.number,
                                          hint: '5',
                                        ),
                                      ),
                                      const SizedBox(width: 20),
                                      Expanded(
                                        child: _buildFullscreenTextField(
                                          controller: _bigBlindController,
                                          label: 'Big Blind',
                                          icon: Icons.casino,
                                          keyboardType: TextInputType.number,
                                          hint: '10',
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 24),

                                  // Max Players
                                  _buildFullscreenTextField(
                                    controller: _maxPlayersController,
                                    label: 'Maximum Players',
                                    icon: Icons.people,
                                    keyboardType: TextInputType.number,
                                    hint: '2-9 players',
                                  ),
                                  const SizedBox(height: 32),

                                  // Game Info Panel
                                  Container(
                                    width: double.infinity,
                                    padding: const EdgeInsets.all(20),
                                    decoration: BoxDecoration(
                                      color: const Color(
                                        0xFF059669,
                                      ).withOpacity(0.1),
                                      borderRadius: BorderRadius.circular(16),
                                      border: Border.all(
                                        color: const Color(
                                          0xFF059669,
                                        ).withOpacity(0.3),
                                        width: 1.5,
                                      ),
                                    ),
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Row(
                                          children: [
                                            Container(
                                              padding: const EdgeInsets.all(8),
                                              decoration: BoxDecoration(
                                                color: const Color(
                                                  0xFF059669,
                                                ).withOpacity(0.2),
                                                borderRadius:
                                                    BorderRadius.circular(8),
                                              ),
                                              child: const Icon(
                                                Icons.info_outline,
                                                color: Color(0xFF059669),
                                                size: 20,
                                              ),
                                            ),
                                            const SizedBox(width: 12),
                                            const Text(
                                              'Game Settings',
                                              style: TextStyle(
                                                color: Color(0xFF059669),
                                                fontSize: 18,
                                                fontWeight: FontWeight.bold,
                                              ),
                                            ),
                                          ],
                                        ),
                                        const SizedBox(height: 16),
                                        _buildInfoRow(
                                          'Game Type:',
                                          'Texas Hold\'em',
                                        ),
                                        _buildInfoRow('Rake Percentage:', '5%'),
                                        _buildInfoRow('Maximum Rake:', '\$50'),
                                        _buildInfoRow(
                                          'Table Visibility:',
                                          'Public',
                                        ),
                                        _buildInfoRow(
                                          'Auto-Start:',
                                          'When 2+ players',
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              );
                            }
                          },
                        ),
                      ),
                    ),
                  ),
                ),
                // Footer
                Container(
                  padding: const EdgeInsets.all(24),
                  decoration: BoxDecoration(
                    color: const Color(0xFF111827),
                    border: Border(
                      top: BorderSide(
                        color: const Color(0xFF059669).withOpacity(0.3),
                        width: 1,
                      ),
                    ),
                  ),
                  child: SafeArea(
                    top: false,
                    child: Row(
                      children: [
                        Expanded(
                          child: TextButton(
                            onPressed: () => Navigator.of(context).pop(),
                            style: TextButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 16),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                                side: BorderSide(color: Colors.grey.shade600),
                              ),
                            ),
                            child: const Text(
                              'Cancel',
                              style: TextStyle(
                                color: Colors.grey,
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 16),
                        Expanded(
                          flex: 2,
                          child: ElevatedButton(
                            onPressed: () {
                              // Validate table name first
                              if (_nameController.text.trim().isEmpty) {
                                ScaffoldMessenger.of(context).showSnackBar(
                                  const SnackBar(
                                    content: Text('Please enter a table name'),
                                    backgroundColor: Colors.red,
                                  ),
                                );
                                return;
                              }

                              // All good, proceed with creation
                              _createTable(
                                _nameController.text,
                                int.tryParse(_minBuyInController.text) ?? 100,
                                int.tryParse(_maxBuyInController.text) ?? 1000,
                                int.tryParse(_smallBlindController.text) ?? 5,
                                int.tryParse(_bigBlindController.text) ?? 10,
                                int.tryParse(_maxPlayersController.text) ?? 9,
                              );
                            },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: const Color(0xFF059669),
                              foregroundColor: Colors.white,
                              padding: const EdgeInsets.symmetric(vertical: 16),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(12),
                              ),
                              elevation: 4,
                              shadowColor: const Color(
                                0xFF059669,
                              ).withOpacity(0.5),
                            ),
                            child: const Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(Icons.add_circle_outline, size: 20),
                                SizedBox(width: 8),
                                Text(
                                  'Create Table',
                                  style: TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  Widget _buildFullscreenTextField({
    required TextEditingController controller,
    required String label,
    required IconData icon,
    String? hint,
    TextInputType? keyboardType,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: const TextStyle(
            color: Color(0xFF9CA3AF),
            fontSize: 16,
            fontWeight: FontWeight.w600,
          ),
        ),
        const SizedBox(height: 8),
        Container(
          height: 56,
          decoration: BoxDecoration(
            gradient: const LinearGradient(
              colors: [Color(0xFF374151), Color(0xFF1F2937)],
            ),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: const Color(0xFF059669).withOpacity(0.3),
              width: 1.5,
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.2),
                blurRadius: 8,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: TextField(
            controller: controller,
            keyboardType: keyboardType,
            textAlignVertical: TextAlignVertical.center,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 16,
              fontWeight: FontWeight.w500,
            ),
            decoration: InputDecoration(
              border: InputBorder.none,
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 16,
                vertical: 16,
              ),
              prefixIcon: Container(
                margin: const EdgeInsets.all(8),
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: const Color(0xFF059669).withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Icon(icon, color: const Color(0xFF059669), size: 20),
              ),
              hintText: hint,
              hintStyle: TextStyle(color: Colors.grey.shade500, fontSize: 16),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            label,
            style: TextStyle(color: Colors.grey.shade400, fontSize: 11),
          ),
          Text(
            value,
            style: const TextStyle(
              color: Colors.white,
              fontSize: 11,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCompactInfoGrid(List<List<String>> items) {
    return Column(
      children: [
        for (int i = 0; i < items.length; i += 2)
          Padding(
            padding: const EdgeInsets.only(bottom: 6),
            child: Row(
              children: [
                Expanded(
                  child: _buildCompactInfoItem(items[i][0], items[i][1]),
                ),
                if (i + 1 < items.length) ...[
                  const SizedBox(width: 12),
                  Expanded(
                    child: _buildCompactInfoItem(
                      items[i + 1][0],
                      items[i + 1][1],
                    ),
                  ),
                ],
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildCompactInfoItem(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          label,
          style: TextStyle(
            color: Colors.grey.shade500,
            fontSize: 11,
            fontWeight: FontWeight.w500,
          ),
        ),
        const SizedBox(height: 2),
        Text(
          value,
          style: const TextStyle(
            color: Color(0xFF059669),
            fontSize: 12,
            fontWeight: FontWeight.w600,
          ),
        ),
      ],
    );
  }

  void _createTable(
    String name,
    int minBuyIn,
    int maxBuyIn,
    int smallBlind,
    int bigBlind,
    int maxPlayers,
  ) {
    // Additional business rule validations
    if (maxBuyIn <= minBuyIn) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Max buy-in must be greater than min buy-in'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    if (bigBlind <= smallBlind) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Big blind must be greater than small blind'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    if (maxPlayers < 2 || maxPlayers > 9) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Max players must be between 2 and 9'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    // Close the dialog
    Navigator.of(context).pop();

    // Send create table message via WebSocket
    final pokerNotifier = ref.read(pokerProvider.notifier);
    pokerNotifier.createTable(
      name: name.trim(),
      minBuyIn: minBuyIn,
      maxBuyIn: maxBuyIn,
      smallBlind: smallBlind,
      bigBlind: bigBlind,
      maxPlayers: maxPlayers,
    );

    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Creating table...'),
        backgroundColor: Color(0xFF059669),
      ),
    );
  }
}
