import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../models/poker_models.dart';
import '../providers/poker_provider.dart';
import '../widgets/poker_widgets.dart';

class PokerTableScreen extends ConsumerStatefulWidget {
  const PokerTableScreen({super.key});

  @override
  ConsumerState<PokerTableScreen> createState() => _PokerTableScreenState();
}

class _PokerTableScreenState extends ConsumerState<PokerTableScreen> {
  @override
  void initState() {
    super.initState();
    // Load game state when screen initializes
    WidgetsBinding.instance.addPostFrameCallback((_) {
      ref.read(gameStateProvider.notifier).refresh();
    });
  }

  @override
  Widget build(BuildContext context) {
    final currentTable = ref.watch(currentTableProvider);
    final gameState = ref.watch(gameStateProvider);
    final currentPlayer = ref.watch(currentPlayerProvider);
    final isMyTurn = ref.watch(isMyTurnProvider);
    final availableActions = ref.watch(availableActionsProvider);

    // Listen to poker events
    ref.listen<AsyncValue<Map<String, dynamic>>>(pokerEventsProvider, (
      previous,
      next,
    ) {
      next.whenData((event) {
        _handlePokerEvent(event);
      });
    });

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

    return Scaffold(
      appBar: AppBar(
        title: Text(currentTable.value?.name ?? 'Poker Table'),
        backgroundColor: Colors.green[700],
        foregroundColor: Colors.white,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            ref.read(currentTableProvider.notifier).leaveTable();
            Navigator.of(context).pop();
          },
        ),
        actions: [
          if (currentTable.value != null && gameState.value != null)
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () => ref.read(gameStateProvider.notifier).refresh(),
            ),
        ],
      ),
      body: currentTable.when(
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
                onPressed: () => Navigator.of(context).pop(),
                child: const Text('Back to Lobby'),
              ),
            ],
          ),
        ),
        data: (table) {
          if (table == null) {
            return const Center(child: Text('No table found'));
          }

          return gameState.when(
            loading: () => const Center(child: CircularProgressIndicator()),
            error: (error, stack) => Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const Icon(Icons.error, size: 64, color: Colors.red),
                  const SizedBox(height: 16),
                  Text('Game Error: $error'),
                  const SizedBox(height: 16),
                  ElevatedButton(
                    onPressed: () =>
                        ref.read(gameStateProvider.notifier).refresh(),
                    child: const Text('Retry'),
                  ),
                ],
              ),
            ),
            data: (game) => _buildTableView(
              context,
              table,
              game,
              currentPlayer,
              isMyTurn,
              availableActions,
            ),
          );
        },
      ),
    );
  }

  Widget _buildTableView(
    BuildContext context,
    PokerTable table,
    PokerGameState? game,
    PokerPlayer? currentPlayer,
    bool isMyTurn,
    List<PokerActionType> availableActions,
  ) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          colors: [Colors.green[800]!, Colors.green[600]!],
        ),
      ),
      child: SafeArea(
        child: Column(
          children: [
            // Table info bar
            Container(
              width: double.infinity,
              padding: const EdgeInsets.all(16),
              color: Colors.black.withOpacity(0.2),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Column(
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
                      Text(
                        'Players: ${table.currentPlayers}/${table.maxPlayers}',
                        style: const TextStyle(
                          color: Colors.white70,
                          fontSize: 14,
                        ),
                      ),
                    ],
                  ),
                  if (game != null)
                    Column(
                      crossAxisAlignment: CrossAxisAlignment.end,
                      children: [
                        Text(
                          'Phase: ${game.phase.toUpperCase()}',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 14,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        Text(
                          'Blinds: \$${game.smallBlind}/\$${game.bigBlind}',
                          style: const TextStyle(
                            color: Colors.white70,
                            fontSize: 12,
                          ),
                        ),
                      ],
                    ),
                ],
              ),
            ),

            // Game area
            Expanded(
              child: game == null
                  ? _buildWaitingArea(table, currentPlayer)
                  : _buildGameArea(
                      game,
                      currentPlayer,
                      isMyTurn,
                      availableActions,
                    ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildWaitingArea(PokerTable table, PokerPlayer? currentPlayer) {
    return Center(
      child: Container(
        padding: const EdgeInsets.all(32),
        margin: const EdgeInsets.all(32),
        decoration: BoxDecoration(
          color: Colors.white,
          borderRadius: BorderRadius.circular(16),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.2),
              blurRadius: 8,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.people, size: 64, color: Colors.green[600]),
            const SizedBox(height: 16),
            const Text(
              'Waiting for Game to Start',
              style: TextStyle(fontSize: 24, fontWeight: FontWeight.bold),
            ),
            const SizedBox(height: 8),
            Text(
              'Players: ${table.currentPlayers}/${table.maxPlayers}',
              style: TextStyle(fontSize: 16, color: Colors.grey[600]),
            ),
            const SizedBox(height: 24),
            if (currentPlayer != null) ...[
              if (!currentPlayer.isReady) ...[
                ElevatedButton.icon(
                  onPressed: () =>
                      ref.read(currentTableProvider.notifier).setReady(true),
                  icon: const Icon(Icons.check),
                  label: const Text('Ready'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green[600],
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 32,
                      vertical: 16,
                    ),
                  ),
                ),
              ] else ...[
                ElevatedButton.icon(
                  onPressed: () =>
                      ref.read(currentTableProvider.notifier).setReady(false),
                  icon: const Icon(Icons.pause),
                  label: const Text('Not Ready'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.orange[600],
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 32,
                      vertical: 16,
                    ),
                  ),
                ),
                const SizedBox(height: 8),
                const Text(
                  'Waiting for other players...',
                  style: TextStyle(color: Colors.grey),
                ),
              ],
              const SizedBox(height: 16),
              // Show start game button if user is table creator and has enough players
              if (table.currentPlayers >= 2) ...[
                ElevatedButton.icon(
                  onPressed: () =>
                      ref.read(currentTableProvider.notifier).startGame(),
                  icon: const Icon(Icons.play_arrow),
                  label: const Text('Start Game'),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blue[600],
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 32,
                      vertical: 16,
                    ),
                  ),
                ),
              ],
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildGameArea(
    PokerGameState game,
    PokerPlayer? currentPlayer,
    bool isMyTurn,
    List<PokerActionType> availableActions,
  ) {
    return Column(
      children: [
        // Pot and community cards area
        Expanded(
          flex: 2,
          child: Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                // Pot
                PotWidget(pot: game.pot, currentBet: game.currentBet),
                const SizedBox(height: 20),
                // Community cards
                CommunityCardsWidget(cards: game.communityCards),
              ],
            ),
          ),
        ),

        // Players area
        Expanded(flex: 2, child: _buildPlayersArea(game, currentPlayer)),

        // Action buttons (for current player)
        if (currentPlayer != null && isMyTurn && availableActions.isNotEmpty)
          Container(
            padding: const EdgeInsets.all(16),
            color: Colors.black.withOpacity(0.1),
            child: _buildActionButtons(availableActions),
          ),

        // Player's hand
        if (currentPlayer != null && currentPlayer.holeCards.isNotEmpty)
          Container(
            padding: const EdgeInsets.all(16),
            child: Column(
              children: [
                const Text(
                  'Your Hand',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                const SizedBox(height: 8),
                HandWidget(
                  cards: currentPlayer.holeCards,
                  cardWidth: 80,
                  cardHeight: 112,
                ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildPlayersArea(PokerGameState game, PokerPlayer? currentPlayer) {
    return Container(
      padding: const EdgeInsets.all(16),
      child: Wrap(
        alignment: WrapAlignment.center,
        spacing: 12,
        runSpacing: 12,
        children: game.players.map((player) {
          return PlayerWidget(
            player: player,
            isActive: game.activePlayerID == player.id,
            isDealer: game.dealerPosition == player.position,
            isSmallBlind: game.smallBlindPosition == player.position,
            isBigBlind: game.bigBlindPosition == player.position,
            isCurrentUser: currentPlayer?.id == player.id,
          );
        }).toList(),
      ),
    );
  }

  Widget _buildActionButtons(List<PokerActionType> availableActions) {
    return Wrap(
      alignment: WrapAlignment.center,
      spacing: 8,
      runSpacing: 8,
      children: availableActions.map((action) {
        return _buildActionButton(action);
      }).toList(),
    );
  }

  Widget _buildActionButton(PokerActionType action) {
    final gameStateNotifier = ref.read(gameStateProvider.notifier);

    switch (action) {
      case PokerActionType.fold:
        return ElevatedButton(
          onPressed: () => gameStateNotifier.fold(),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.red[600]),
          child: const Text('Fold', style: TextStyle(color: Colors.white)),
        );
      case PokerActionType.check:
        return ElevatedButton(
          onPressed: () => gameStateNotifier.check(),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.orange[600]),
          child: const Text('Check', style: TextStyle(color: Colors.white)),
        );
      case PokerActionType.call:
        final callAmount = gameStateNotifier.getCallAmount();
        return ElevatedButton(
          onPressed: () => gameStateNotifier.call(),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.blue[600]),
          child: Text(
            'Call \$${callAmount}',
            style: const TextStyle(color: Colors.white),
          ),
        );
      case PokerActionType.bet:
        return ElevatedButton(
          onPressed: () => _showBetDialog(PokerActionType.bet),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.green[600]),
          child: const Text('Bet', style: TextStyle(color: Colors.white)),
        );
      case PokerActionType.raise:
        return ElevatedButton(
          onPressed: () => _showBetDialog(PokerActionType.raise),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.green[700]),
          child: const Text('Raise', style: TextStyle(color: Colors.white)),
        );
      case PokerActionType.allIn:
        return ElevatedButton(
          onPressed: () => gameStateNotifier.allIn(),
          style: ElevatedButton.styleFrom(backgroundColor: Colors.purple[600]),
          child: const Text('All-In', style: TextStyle(color: Colors.white)),
        );
    }
  }

  void _showBetDialog(PokerActionType action) {
    final gameStateNotifier = ref.read(gameStateProvider.notifier);
    final minAmount = action == PokerActionType.raise
        ? gameStateNotifier.getMinRaiseAmount()
        : gameStateNotifier.getCallAmount() + 1;

    showDialog(
      context: context,
      builder: (context) => BetDialog(
        action: action,
        minAmount: minAmount,
        onBet: (amount) {
          if (action == PokerActionType.bet) {
            gameStateNotifier.bet(amount);
          } else {
            gameStateNotifier.raise(amount);
          }
        },
      ),
    );
  }

  void _handlePokerEvent(Map<String, dynamic> event) {
    final type = event['type'] as String;

    switch (type) {
      case 'game_started':
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Game started!'),
              backgroundColor: Colors.green,
            ),
          );
        }
        break;
      case 'game_ended':
        if (mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(
              content: Text('Game ended!'),
              backgroundColor: Colors.blue,
            ),
          );
        }
        break;
      case 'player_joined':
      case 'player_left':
        // Refresh game state
        ref.read(gameStateProvider.notifier).refresh();
        break;
    }
  }
}

class BetDialog extends StatefulWidget {
  final PokerActionType action;
  final int minAmount;
  final Function(int) onBet;

  const BetDialog({
    super.key,
    required this.action,
    required this.minAmount,
    required this.onBet,
  });

  @override
  State<BetDialog> createState() => _BetDialogState();
}

class _BetDialogState extends State<BetDialog> {
  late int _amount;
  final _controller = TextEditingController();

  @override
  void initState() {
    super.initState();
    _amount = widget.minAmount;
    _controller.text = _amount.toString();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text('${widget.action.name.toUpperCase()} Amount'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          TextField(
            controller: _controller,
            keyboardType: TextInputType.number,
            decoration: InputDecoration(
              labelText: 'Amount',
              prefixText: '\$',
              hintText: 'Enter amount',
              border: const OutlineInputBorder(),
              helperText: 'Minimum: \$${widget.minAmount}',
            ),
            onChanged: (value) {
              final amount = int.tryParse(value);
              if (amount != null && amount >= widget.minAmount) {
                setState(() {
                  _amount = amount;
                });
              }
            },
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: _amount >= widget.minAmount
              ? () {
                  widget.onBet(_amount);
                  Navigator.of(context).pop();
                }
              : null,
          child: Text(widget.action.name.toUpperCase()),
        ),
      ],
    );
  }
}
