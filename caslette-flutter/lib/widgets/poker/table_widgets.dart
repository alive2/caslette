import 'package:flutter/material.dart';
import '../../models/poker/poker_models.dart' as poker;
import 'card_widgets.dart';
import 'player_widgets.dart';

class PokerTableLayout extends StatelessWidget {
  final poker.PokerTable table;
  final String? currentUserId;
  final Function(int seatNumber)? onJoinSeat;
  final Function(poker.PlayerAction action, {int? amount})? onPlayerAction;

  const PokerTableLayout({
    super.key,
    required this.table,
    this.currentUserId,
    this.onJoinSeat,
    this.onPlayerAction,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: const RadialGradient(
          center: Alignment.center,
          radius: 1.5,
          colors: [
            Color(0xFF059669), // Emerald center
            Color(0xFF064E3B), // Dark emerald edge
          ],
        ),
        borderRadius: BorderRadius.circular(20),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.3),
            blurRadius: 20,
            offset: const Offset(0, 8),
          ),
        ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          children: [
            // Table info header
            _buildTableHeader(),
            const SizedBox(height: 16),

            // Main table area
            Expanded(child: _buildTableArea()),

            // Player controls (if applicable)
            if (_shouldShowControls())
              Column(
                children: [const SizedBox(height: 16), _buildPlayerControls()],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildTableHeader() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.2),
        borderRadius: BorderRadius.circular(12),
      ),
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
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                'Blinds: ${table.smallBlind}/${table.bigBlind}',
                style: TextStyle(
                  color: Colors.white.withOpacity(0.8),
                  fontSize: 12,
                ),
              ),
            ],
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              Text(
                'Players: ${table.currentPlayers}/${table.maxPlayers}',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 14,
                  fontWeight: FontWeight.w500,
                ),
              ),
              Text(
                'Hand #${table.handNumber}',
                style: TextStyle(
                  color: Colors.white.withOpacity(0.8),
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildTableArea() {
    return Stack(
      children: [
        // Table oval
        Center(
          child: Container(
            width: 320,
            height: 180,
            decoration: BoxDecoration(
              color: const Color(0xFF065F46),
              borderRadius: BorderRadius.circular(90),
              border: Border.all(color: const Color(0xFFF59E0B), width: 3),
            ),
          ),
        ),

        // Center area with community cards and pot
        Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              // Community cards
              if (table.communityCards.isNotEmpty)
                CommunityCards(
                  cards: table.communityCards,
                  bettingRound: table.bettingRound,
                  cardWidth: 40,
                  cardHeight: 56,
                ),

              const SizedBox(height: 12),

              // Pot information
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: 12,
                  vertical: 6,
                ),
                decoration: BoxDecoration(
                  color: const Color(0xFFF59E0B).withOpacity(0.2),
                  borderRadius: BorderRadius.circular(16),
                  border: Border.all(color: const Color(0xFFF59E0B), width: 1),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(
                      Icons.monetization_on,
                      color: Color(0xFFF59E0B),
                      size: 16,
                    ),
                    const SizedBox(width: 4),
                    Text(
                      'Pot: ${table.pot}',
                      style: const TextStyle(
                        color: Color(0xFFF59E0B),
                        fontSize: 14,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ),

              // Betting round indicator
              Container(
                margin: const EdgeInsets.only(top: 8),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: const Color(0xFF9333EA).withOpacity(0.2),
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Text(
                  table.bettingRound.displayName,
                  style: const TextStyle(
                    color: Color(0xFF9333EA),
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),
            ],
          ),
        ),

        // Player seats arranged around the table
        ..._buildPlayerSeats(),
      ],
    );
  }

  List<Widget> _buildPlayerSeats() {
    final seats = <Widget>[];

    // Define seat positions around the table (8 seats)
    final seatPositions = [
      const Alignment(0, -0.8), // Top center
      const Alignment(0.6, -0.6), // Top right
      const Alignment(0.8, 0), // Right
      const Alignment(0.6, 0.6), // Bottom right
      const Alignment(0, 0.8), // Bottom center
      const Alignment(-0.6, 0.6), // Bottom left
      const Alignment(-0.8, 0), // Left
      const Alignment(-0.6, -0.6), // Top left
    ];

    for (int i = 0; i < table.maxPlayers && i < seatPositions.length; i++) {
      final seatNumber = i + 1;
      final player = table.players
          .where((p) => p.seatNumber == seatNumber)
          .firstOrNull;
      final isEmpty = player == null;
      final canJoin =
          isEmpty && table.hasAvailableSeats && currentUserId != null;
      final isCurrentPlayer = player?.userId == table.currentPlayerUserId;

      seats.add(
        Align(
          alignment: seatPositions[i],
          child: PlayerSeat(
            player: player,
            isEmpty: isEmpty,
            isCurrentPlayer: isCurrentPlayer,
            canJoin: canJoin,
            seatNumber: seatNumber,
            currentUserId: currentUserId,
            onJoinSeat: canJoin ? () => onJoinSeat?.call(seatNumber) : null,
          ),
        ),
      );
    }

    return seats;
  }

  bool _shouldShowControls() {
    if (currentUserId == null || onPlayerAction == null) return false;

    final currentPlayer = table.getPlayerByUserId(currentUserId!);
    if (currentPlayer == null) return false;

    return table.currentPlayerUserId == currentUserId &&
        table.status == poker.TableStatus.active;
  }

  Widget _buildPlayerControls() {
    final currentPlayer = table.getPlayerByUserId(currentUserId!);
    if (currentPlayer == null) return const SizedBox();

    // Determine available actions based on game state
    final canFold = currentPlayer.status == poker.PlayerStatus.active;
    final canCheck =
        table.currentBet == 0 || currentPlayer.currentBet == table.currentBet;
    final canCall =
        table.currentBet > currentPlayer.currentBet && table.currentBet > 0;
    final canBet = table.currentBet == 0;
    final canRaise =
        table.currentBet > 0 && currentPlayer.currentBet < table.currentBet;

    final callAmount = table.currentBet - currentPlayer.currentBet;
    final minRaise = table.currentBet > 0
        ? table.currentBet * 2
        : table.bigBlind;
    final maxRaise = currentPlayer.chipCount;

    return BettingControls(
      currentBet: callAmount,
      playerChips: currentPlayer.chipCount,
      minRaise: minRaise,
      maxRaise: maxRaise,
      canCheck: canCheck,
      canCall: canCall,
      canBet: canBet,
      canRaise: canRaise,
      canFold: canFold,
      onAction: onPlayerAction!,
    );
  }
}
