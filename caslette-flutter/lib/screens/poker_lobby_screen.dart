import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/poker_models.dart';
import '../providers/poker_provider.dart';
import '../providers/auth_provider.dart';
import 'create_table_dialog.dart';
import 'poker_table_screen.dart';

class PokerLobbyScreen extends ConsumerStatefulWidget {
  const PokerLobbyScreen({super.key});

  @override
  ConsumerState<PokerLobbyScreen> createState() => _PokerLobbyScreenState();
}

class _PokerLobbyScreenState extends ConsumerState<PokerLobbyScreen> {
  @override
  void initState() {
    super.initState();
    // Load tables when screen initializes
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(pokerTablesProvider.notifier).refresh();
    });
  }

  @override
  Widget build(BuildContext context) {
    final tablesAsync = ref.watch(pokerTablesProvider);
    final isInTable = ref.watch(isInTableProvider);

    // Listen to poker errors
    ref.listen<AsyncValue<String>>(pokerErrorsProvider, (previous, next) {
      next.whenData((error) {
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(content: Text(error), backgroundColor: Colors.red),
          );
        }
      });
    });

    // If already in a table, navigate to table screen
    if (isInTable) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        Navigator.of(context).pushReplacement(
          MaterialPageRoute(builder: (context) => const PokerTableScreen()),
        );
      });
    }

    return Scaffold(
      appBar: AppBar(
        title: const Text('Poker Lobby'),
        backgroundColor: Colors.green[700],
        foregroundColor: Colors.white,
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: () => ref.read(pokerTablesProvider.notifier).refresh(),
          ),
        ],
      ),
      body: tablesAsync.when(
        loading: () => const Center(child: CircularProgressIndicator()),
        error: (error, stack) => Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.error, size: 64, color: Colors.red[400]),
              const SizedBox(height: 16),
              Text('Error: $error'),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () =>
                    ref.read(pokerTablesProvider.notifier).refresh(),
                child: const Text('Retry'),
              ),
            ],
          ),
        ),
        data: (tables) => RefreshIndicator(
          onRefresh: () => ref.read(pokerTablesProvider.notifier).refresh(),
          child: Column(
            children: [
              // Header stats
              Container(
                width: double.infinity,
                padding: const EdgeInsets.all(16),
                color: Colors.green[50],
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceAround,
                  children: [
                    _buildStatCard(
                      'Total Tables',
                      tables.length.toString(),
                      Icons.table_restaurant,
                    ),
                    _buildStatCard(
                      'Playing',
                      tables.where((t) => t.isPlaying).length.toString(),
                      Icons.play_circle,
                    ),
                    _buildStatCard(
                      'Open',
                      tables.where((t) => t.canJoin).length.toString(),
                      Icons.door_front_door,
                    ),
                  ],
                ),
              ),

              // Tables list
              Expanded(
                child: tables.isEmpty
                    ? Center(
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(
                              Icons.table_restaurant,
                              size: 64,
                              color: Colors.grey[400],
                            ),
                            const SizedBox(height: 16),
                            Text(
                              'No tables available',
                              style: TextStyle(
                                fontSize: 18,
                                color: Colors.grey[600],
                              ),
                            ),
                            const SizedBox(height: 8),
                            Text(
                              'Create a new table to get started!',
                              style: TextStyle(color: Colors.grey[500]),
                            ),
                          ],
                        ),
                      )
                    : ListView.builder(
                        padding: const EdgeInsets.all(16),
                        itemCount: tables.length,
                        itemBuilder: (context, index) {
                          final table = tables[index];
                          return _buildTableCard(context, table);
                        },
                      ),
              ),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: () => _showCreateTableDialog(context),
        backgroundColor: Colors.green[600],
        icon: const Icon(Icons.add),
        label: const Text('Create Table'),
      ),
    );
  }

  Widget _buildStatCard(String title, String value, IconData icon) {
    return Container(
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.1),
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: Colors.green[600], size: 24),
          const SizedBox(height: 4),
          Text(
            value,
            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
          ),
          Text(title, style: TextStyle(fontSize: 12, color: Colors.grey[600])),
        ],
      ),
    );
  }

  Widget _buildTableCard(BuildContext context, PokerTable table) {
    final settings = table.settings;
    final smallBlind = settings['small_blind'] ?? 10;
    final bigBlind = settings['big_blind'] ?? 20;

    return Material(
      child: InkWell(
        onTap: table.canJoin ? () => _joinTable(table.id) : null,
        borderRadius: BorderRadius.circular(8),
        child: Container(
          margin: const EdgeInsets.only(bottom: 12),
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(8),
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.1),
                blurRadius: 4,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(
                      table.name,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Row(
                    children: [
                      if (table.isPrivate)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 8,
                            vertical: 4,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.orange[100],
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(
                                Icons.lock,
                                size: 12,
                                color: Colors.orange[700],
                              ),
                              const SizedBox(width: 4),
                              Text(
                                'Private',
                                style: TextStyle(
                                  fontSize: 10,
                                  color: Colors.orange[700],
                                  fontWeight: FontWeight.w600,
                                ),
                              ),
                            ],
                          ),
                        ),
                      const SizedBox(width: 8),
                      Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 4,
                        ),
                        decoration: BoxDecoration(
                          color: _getStatusColor(table.status).withOpacity(0.1),
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Text(
                          table.status.toUpperCase(),
                          style: TextStyle(
                            fontSize: 10,
                            color: _getStatusColor(table.status),
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
              const SizedBox(height: 8),

              Row(
                children: [
                  _buildTableInfo(Icons.casino, 'Texas Hold\'em'),
                  const SizedBox(width: 16),
                  _buildTableInfo(
                    Icons.monetization_on,
                    '\$$smallBlind/\$$bigBlind',
                  ),
                  const SizedBox(width: 16),
                  _buildTableInfo(
                    Icons.people,
                    '${table.currentPlayers}/${table.maxPlayers}',
                  ),
                ],
              ),

              if (table.observerCount > 0) ...[
                const SizedBox(height: 4),
                Row(
                  children: [
                    _buildTableInfo(
                      Icons.visibility,
                      '${table.observerCount} watching',
                    ),
                  ],
                ),
              ],

              const SizedBox(height: 12),

              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    'Created: ${_formatDateTime(table.createdAt)}',
                    style: TextStyle(fontSize: 12, color: Colors.grey[600]),
                  ),
                  Row(
                    children: [
                      if (!table.canJoin &&
                          !table.isFull &&
                          !_isCurrentUserAtTable(table))
                        ElevatedButton.icon(
                          onPressed: () => _joinTableAsObserver(table.id),
                          icon: const Icon(Icons.visibility, size: 16),
                          label: const Text('Watch'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.grey[600],
                            foregroundColor: Colors.white,
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 6,
                            ),
                            minimumSize: Size.zero,
                          ),
                        ),
                      if (_isCurrentUserAtTable(table)) ...[
                        const SizedBox(width: 8),
                        ElevatedButton.icon(
                          onPressed: () => _navigateToTable(table.id),
                          icon: const Icon(Icons.play_arrow, size: 16),
                          label: const Text('Enter'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.blue[600],
                            foregroundColor: Colors.white,
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 6,
                            ),
                            minimumSize: Size.zero,
                          ),
                        ),
                      ] else if (table.canJoin) ...[
                        const SizedBox(width: 8),
                        ElevatedButton.icon(
                          onPressed: () => _joinTable(table.id),
                          icon: const Icon(Icons.play_arrow, size: 16),
                          label: const Text('Join'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.green[600],
                            foregroundColor: Colors.white,
                            padding: const EdgeInsets.symmetric(
                              horizontal: 12,
                              vertical: 6,
                            ),
                            minimumSize: Size.zero,
                          ),
                        ),
                      ],
                      if (table.isFull)
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: Colors.red[100],
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Text(
                            'FULL',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.red[700],
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                    ],
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildTableInfo(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: Colors.grey[600]),
        const SizedBox(width: 4),
        Text(text, style: TextStyle(fontSize: 12, color: Colors.grey[700])),
      ],
    );
  }

  Color _getStatusColor(String status) {
    switch (status.toLowerCase()) {
      case 'waiting':
        return Colors.orange;
      case 'playing':
        return Colors.green;
      case 'finished':
        return Colors.grey;
      default:
        return Colors.blue;
    }
  }

  String _formatDateTime(DateTime dateTime) {
    final now = DateTime.now();
    final difference = now.difference(dateTime);

    if (difference.inMinutes < 1) {
      return 'Just now';
    } else if (difference.inHours < 1) {
      return '${difference.inMinutes}m ago';
    } else if (difference.inDays < 1) {
      return '${difference.inHours}h ago';
    } else {
      return '${dateTime.day}/${dateTime.month}';
    }
  }

  void _showCreateTableDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (context) => const CreateTableDialog(),
    );
  }

  void _joinTable(String tableId) {
    ref.read(currentTableProvider.notifier).joinTable(tableId);
  }

  void _joinTableAsObserver(String tableId) {
    ref
        .read(currentTableProvider.notifier)
        .joinTable(tableId, mode: 'observer');
  }

  bool _isCurrentUserAtTable(PokerTable table) {
    final currentUser = ref.read(authProvider.notifier).currentUser;
    if (currentUser == null) return false;

    return table.playerSlots.any(
      (slot) => slot.playerId != null && slot.playerId == currentUser.id,
    );
  }

  void _navigateToTable(String tableId) {
    // First join the table (which sets it as current table) then navigate
    ref
        .read(currentTableProvider.notifier)
        .joinTable(tableId)
        .then((_) {
          Navigator.of(context).pushNamed('/poker/table');
        })
        .catchError((error) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text('Failed to enter table: $error'),
              backgroundColor: Colors.red,
            ),
          );
        });
  }
}
