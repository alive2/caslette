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
    final screenSize = MediaQuery.of(context).size;
    final isLandscape = screenSize.width > screenSize.height;

    return Container(
      width: double.infinity,
      height: double.infinity,
      decoration: const BoxDecoration(
        image: DecorationImage(
          image: AssetImage('assets/images/poker_table_texture.png'),
          fit: BoxFit.cover,
          scale: 0.5,
        ),
      ),
      child: Padding(
        padding: EdgeInsets.all(isLandscape ? 12 : 8),
        child: Column(
          children: [
            // Main table area (takes full screen)
            Expanded(child: _buildFullscreenTableArea()),

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

  Widget _buildFullscreenTableArea() {
    return LayoutBuilder(
      builder: (context, constraints) {
        // Calculate responsive table size
        final screenWidth = constraints.maxWidth;
        final screenHeight = constraints.maxHeight;
        final tableWidth = (screenWidth * 0.8).clamp(280.0, 500.0);
        final tableHeight = (screenHeight * 0.5).clamp(160.0, 280.0);

        return Stack(
          children: [
            // Transparent table overlay for content positioning
            Center(
              child: Container(
                width: tableWidth,
                height: tableHeight,
                decoration: BoxDecoration(
                  color: Colors.transparent,
                  borderRadius: BorderRadius.circular(tableHeight / 2),
                  // Optional: subtle border to define the table area
                  border: Border.all(
                    color: Colors.white.withOpacity(0.1),
                    width: 2,
                  ),
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
                      cardWidth: 48,
                      cardHeight: 68,
                    ),

                  const SizedBox(height: 16),

                  // Enhanced pot with gaming effects
                  AnimatedContainer(
                    duration: const Duration(milliseconds: 500),
                    curve: Curves.elasticOut,
                    padding: const EdgeInsets.symmetric(
                      horizontal: 20,
                      vertical: 12,
                    ),
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          const Color(0xFFFFD700).withOpacity(0.9), // Gold
                          const Color(0xFFF59E0B).withOpacity(0.8), // Amber
                          const Color(0xFFD97706).withOpacity(0.7), // Orange
                        ],
                      ),
                      borderRadius: BorderRadius.circular(25),
                      border: Border.all(
                        color: const Color(0xFFFFD700),
                        width: 3,
                      ),
                      boxShadow: [
                        // Outer glow effect
                        BoxShadow(
                          color: const Color(0xFFFFD700).withOpacity(0.6),
                          blurRadius: 20,
                          spreadRadius: 3,
                          offset: const Offset(0, 0),
                        ),
                        // Main shadow
                        BoxShadow(
                          color: Colors.black.withOpacity(0.3),
                          blurRadius: 12,
                          offset: const Offset(0, 6),
                        ),
                        // Inner highlight
                        BoxShadow(
                          color: Colors.white.withOpacity(0.3),
                          blurRadius: 3,
                          offset: const Offset(0, -1),
                        ),
                      ],
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        // Enhanced diamond with glow
                        Container(
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(10),
                            boxShadow: [
                              BoxShadow(
                                color: const Color(0xFFFFD700).withOpacity(0.5),
                                blurRadius: 8,
                                offset: const Offset(0, 0),
                              ),
                            ],
                          ),
                          child: const Icon(
                            Icons.diamond,
                            color: Colors.white,
                            size: 22,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Text(
                          'POT: ${table.pot}',
                          style: const TextStyle(
                            color: Colors.white,
                            fontSize: 18,
                            fontWeight: FontWeight.w900,
                            letterSpacing: 1.2,
                            shadows: [
                              Shadow(
                                color: Colors.black,
                                offset: Offset(1, 1),
                                blurRadius: 2,
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),

                  // Enhanced betting round indicator
                  AnimatedContainer(
                    duration: const Duration(milliseconds: 400),
                    margin: const EdgeInsets.only(top: 12),
                    padding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 8,
                    ),
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          const Color(0xFF7C3AED).withOpacity(0.8),
                          const Color(0xFF9333EA).withOpacity(0.7),
                        ],
                      ),
                      borderRadius: BorderRadius.circular(18),
                      border: Border.all(
                        color: const Color(0xFF7C3AED),
                        width: 2,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFF7C3AED).withOpacity(0.4),
                          blurRadius: 12,
                          offset: const Offset(0, 0),
                        ),
                        BoxShadow(
                          color: Colors.black.withOpacity(0.2),
                          blurRadius: 6,
                          offset: const Offset(0, 3),
                        ),
                      ],
                    ),
                    child: Text(
                      table.bettingRound.displayName.toUpperCase(),
                      style: const TextStyle(
                        color: Colors.white,
                        fontSize: 13,
                        fontWeight: FontWeight.w800,
                        letterSpacing: 1.0,
                        shadows: [
                          Shadow(
                            color: Colors.black,
                            offset: Offset(0, 1),
                            blurRadius: 2,
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ),

            // Player seats arranged around the table
            ..._buildPlayerSeats(
              tableWidth,
              tableHeight,
              constraints,
              currentUserId,
            ),
          ],
        );
      },
    );
  }

  List<Widget> _buildPlayerSeats(
    double tableWidth,
    double tableHeight,
    BoxConstraints constraints,
    String? currentUserId,
  ) {
    final seats = <Widget>[];

    // Create predefined seat positions around the table for better control
    final seatPositions = [
      const Alignment(0, -0.85), // Top center
      const Alignment(0.65, -0.65), // Top right
      const Alignment(0.85, 0), // Right center
      const Alignment(0.65, 0.65), // Bottom right
      const Alignment(0, 0.85), // Bottom center
      const Alignment(-0.65, 0.65), // Bottom left
      const Alignment(-0.85, 0), // Left center
      const Alignment(-0.65, -0.65), // Top left
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
