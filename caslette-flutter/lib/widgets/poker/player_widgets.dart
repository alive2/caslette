import 'package:flutter/material.dart';
import '../../models/poker/poker_models.dart' as poker;
import 'card_widgets.dart';

class PlayerSeat extends StatelessWidget {
  final poker.PokerPlayer? player;
  final bool isEmpty;
  final bool isCurrentPlayer;
  final bool canJoin;
  final int seatNumber;
  final VoidCallback? onJoinSeat;
  final String? currentUserId;

  const PlayerSeat({
    super.key,
    this.player,
    this.isEmpty = false,
    this.isCurrentPlayer = false,
    this.canJoin = false,
    required this.seatNumber,
    this.onJoinSeat,
    this.currentUserId,
  });

  @override
  Widget build(BuildContext context) {
    if (isEmpty && !canJoin) {
      return _buildEmptySeat();
    }

    if (isEmpty && canJoin) {
      return _buildJoinableSeat();
    }

    return _buildOccupiedSeat();
  }

  Widget _buildEmptySeat() {
    return AnimatedContainer(
      duration: const Duration(milliseconds: 200),
      width: 80,
      height: 60,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
          colors: [
            const Color(0xFF374151).withOpacity(0.4),
            const Color(0xFF1F2937).withOpacity(0.6),
          ],
        ),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: const Color(0xFF6B7280).withOpacity(0.6),
          width: 1.5,
        ),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.3),
            blurRadius: 6,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Center(
        child: Text(
          'Seat $seatNumber',
          style: TextStyle(
            color: Colors.white.withOpacity(0.6),
            fontSize: 10,
            fontWeight: FontWeight.w500,
          ),
        ),
      ),
    );
  }

  Widget _buildJoinableSeat() {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: onJoinSeat,
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 200),
          width: 80,
          height: 60,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                const Color(0xFF10B981).withOpacity(0.3),
                const Color(0xFF059669).withOpacity(0.4),
              ],
            ),
            borderRadius: BorderRadius.circular(12),
            border: Border.all(color: const Color(0xFF10B981), width: 2),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF10B981).withOpacity(0.3),
                blurRadius: 8,
                offset: const Offset(0, 0),
              ),
              BoxShadow(
                color: Colors.black.withOpacity(0.2),
                blurRadius: 4,
                offset: const Offset(0, 2),
              ),
            ],
          ),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Container(
                padding: const EdgeInsets.all(3),
                decoration: BoxDecoration(
                  color: Colors.white.withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: const Icon(
                  Icons.add_circle_outline,
                  color: Colors.white,
                  size: 18,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                'Join Seat $seatNumber',
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 9,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildOccupiedSeat() {
    if (player == null) return _buildEmptySeat();

    final isMyPlayer = player!.userId == currentUserId;
    final showCards = isMyPlayer && player!.holeCards.isNotEmpty;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      width: 80,
      decoration: BoxDecoration(
        gradient: isCurrentPlayer
            ? LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [
                  const Color(0xFF7C3AED).withOpacity(0.4),
                  const Color(0xFF9333EA).withOpacity(0.3),
                ],
              )
            : LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [const Color(0xFF374151), const Color(0xFF1F2937)],
              ),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isCurrentPlayer
              ? const Color(0xFF7C3AED)
              : const Color(0xFF4B5563),
          width: isCurrentPlayer ? 2.5 : 1.5,
        ),
        boxShadow: isCurrentPlayer
            ? [
                BoxShadow(
                  color: const Color(0xFF7C3AED).withOpacity(0.5),
                  blurRadius: 12,
                  spreadRadius: 1,
                  offset: const Offset(0, 0),
                ),
                BoxShadow(
                  color: Colors.black.withOpacity(0.3),
                  blurRadius: 6,
                  offset: const Offset(0, 3),
                ),
              ]
            : [
                BoxShadow(
                  color: Colors.black.withOpacity(0.3),
                  blurRadius: 6,
                  offset: const Offset(0, 2),
                ),
              ],
      ),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Player name and status
            Row(
              children: [
                // Enhanced animated dealer button
                if (player!.isDealer)
                  AnimatedContainer(
                    duration: const Duration(milliseconds: 500),
                    width: 18,
                    height: 18,
                    decoration: BoxDecoration(
                      gradient: const LinearGradient(
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                        colors: [
                          Color(0xFFFFD700), // Gold
                          Color(0xFFF59E0B), // Amber
                        ],
                      ),
                      shape: BoxShape.circle,
                      border: Border.all(color: Colors.white, width: 1.5),
                      boxShadow: [
                        BoxShadow(
                          color: const Color(0xFFFFD700).withOpacity(0.6),
                          blurRadius: 8,
                          offset: const Offset(0, 0),
                        ),
                        BoxShadow(
                          color: Colors.black.withOpacity(0.3),
                          blurRadius: 4,
                          offset: const Offset(0, 2),
                        ),
                      ],
                    ),
                    child: const Center(
                      child: Text(
                        'D',
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 11,
                          fontWeight: FontWeight.w900,
                          shadows: [
                            Shadow(
                              color: Colors.black,
                              offset: Offset(0, 1),
                              blurRadius: 1,
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                if (player!.isDealer) const SizedBox(width: 4),

                // Player name
                Expanded(
                  child: Text(
                    player!.username,
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 12,
                      fontWeight: isMyPlayer
                          ? FontWeight.bold
                          : FontWeight.w500,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),

            // Chips
            Row(
              children: [
                const Icon(
                  Icons.monetization_on,
                  color: Color(0xFFF59E0B),
                  size: 14,
                ),
                const SizedBox(width: 4),
                Text(
                  '${player!.chipCount}',
                  style: const TextStyle(
                    color: Color(0xFFF59E0B),
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 4),

            // Current bet (if any)
            if (player!.currentBet > 0)
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                decoration: BoxDecoration(
                  color: const Color(0xFF7C3AED).withOpacity(0.2),
                  borderRadius: BorderRadius.circular(8),
                ),
                child: Text(
                  'Bet: ${player!.currentBet}',
                  style: const TextStyle(
                    color: Color(0xFF7C3AED),
                    fontSize: 10,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ),

            // Player status
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: _getStatusColor().withOpacity(0.2),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(
                player!.status.displayName,
                style: TextStyle(
                  color: _getStatusColor(),
                  fontSize: 10,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ),
            const SizedBox(height: 6),

            // Hole cards (if applicable)
            if (showCards)
              PlayerHoleCards(
                cards: player!.holeCards,
                cardWidth: 30,
                cardHeight: 42,
              ),
          ],
        ),
      ),
    );
  }

  Color _getStatusColor() {
    if (player == null) return Colors.grey;

    switch (player!.status) {
      case poker.PlayerStatus.active:
        return const Color(0xFF059669);
      case poker.PlayerStatus.folded:
        return const Color(0xFF9CA3AF);
      case poker.PlayerStatus.allin:
        return const Color(0xFFDC2626);
      case poker.PlayerStatus.waiting:
        return const Color(0xFFF59E0B);
      case poker.PlayerStatus.sitting_out:
        return const Color(0xFF6B7280);
    }
  }
}

class BettingControls extends StatefulWidget {
  final int currentBet;
  final int playerChips;
  final int minRaise;
  final int maxRaise;
  final bool canCheck;
  final bool canCall;
  final bool canBet;
  final bool canRaise;
  final bool canFold;
  final Function(poker.PlayerAction action, {int? amount}) onAction;

  const BettingControls({
    super.key,
    required this.currentBet,
    required this.playerChips,
    required this.minRaise,
    required this.maxRaise,
    required this.canCheck,
    required this.canCall,
    required this.canBet,
    required this.canRaise,
    required this.canFold,
    required this.onAction,
  });

  @override
  State<BettingControls> createState() => _BettingControlsState();
}

class _BettingControlsState extends State<BettingControls> {
  late TextEditingController _betController;
  int _betAmount = 0;

  @override
  void initState() {
    super.initState();
    _betAmount = widget.minRaise;
    _betController = TextEditingController(text: _betAmount.toString());
  }

  @override
  void dispose() {
    _betController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF1F2937),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: const Color(0xFF374151), width: 1),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Bet amount input
          if (widget.canBet || widget.canRaise)
            Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _betController,
                        keyboardType: TextInputType.number,
                        style: const TextStyle(color: Colors.white),
                        decoration: InputDecoration(
                          labelText: widget.canBet
                              ? 'Bet Amount'
                              : 'Raise Amount',
                          labelStyle: const TextStyle(color: Color(0xFF9CA3AF)),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: const BorderSide(
                              color: Color(0xFF374151),
                            ),
                          ),
                          enabledBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: const BorderSide(
                              color: Color(0xFF374151),
                            ),
                          ),
                          focusedBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: const BorderSide(
                              color: Color(0xFF9333EA),
                            ),
                          ),
                        ),
                        onChanged: (value) {
                          setState(() {
                            _betAmount = int.tryParse(value) ?? widget.minRaise;
                          });
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    ElevatedButton(
                      onPressed: () {
                        setState(() {
                          _betAmount = widget.playerChips;
                          _betController.text = _betAmount.toString();
                        });
                      },
                      style: ElevatedButton.styleFrom(
                        backgroundColor: const Color(0xFFDC2626),
                        foregroundColor: Colors.white,
                        minimumSize: const Size(60, 40),
                      ),
                      child: const Text(
                        'All-in',
                        style: TextStyle(fontSize: 12),
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Text(
                      'Min: ${widget.minRaise}',
                      style: const TextStyle(
                        color: Color(0xFF9CA3AF),
                        fontSize: 12,
                      ),
                    ),
                    const Spacer(),
                    Text(
                      'Max: ${widget.maxRaise}',
                      style: const TextStyle(
                        color: Color(0xFF9CA3AF),
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
              ],
            ),

          // Action buttons
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: [
              if (widget.canFold)
                _buildActionButton(
                  'Fold',
                  Colors.red,
                  () => widget.onAction(poker.PlayerAction.fold),
                ),
              if (widget.canCheck)
                _buildActionButton(
                  'Check',
                  const Color(0xFF059669),
                  () => widget.onAction(poker.PlayerAction.check),
                ),
              if (widget.canCall)
                _buildActionButton(
                  'Call ${widget.currentBet}',
                  const Color(0xFF3B82F6),
                  () => widget.onAction(poker.PlayerAction.call),
                ),
              if (widget.canBet)
                _buildActionButton(
                  'Bet $_betAmount',
                  const Color(0xFF9333EA),
                  () => widget.onAction(
                    poker.PlayerAction.bet,
                    amount: _betAmount,
                  ),
                ),
              if (widget.canRaise)
                _buildActionButton(
                  'Raise $_betAmount',
                  const Color(0xFF9333EA),
                  () => widget.onAction(
                    poker.PlayerAction.raise,
                    amount: _betAmount,
                  ),
                ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildActionButton(String text, Color color, VoidCallback onPressed) {
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 150),
        child: ElevatedButton(
          onPressed: onPressed,
          style:
              ElevatedButton.styleFrom(
                backgroundColor: color,
                foregroundColor: Colors.white,
                minimumSize: const Size(85, 45),
                elevation: 8,
                shadowColor: color.withOpacity(0.4),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ).copyWith(
                overlayColor: MaterialStateProperty.resolveWith<Color?>((
                  Set<MaterialState> states,
                ) {
                  if (states.contains(MaterialState.pressed)) {
                    return Colors.white.withOpacity(0.1);
                  }
                  if (states.contains(MaterialState.hovered)) {
                    return Colors.white.withOpacity(0.05);
                  }
                  return null;
                }),
              ),
          child: Container(
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(12),
              gradient: LinearGradient(
                begin: Alignment.topCenter,
                end: Alignment.bottomCenter,
                colors: [Colors.white.withOpacity(0.2), Colors.transparent],
              ),
            ),
            child: Center(
              child: Text(
                text,
                style: const TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w700,
                  letterSpacing: 0.5,
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
          ),
        ),
      ),
    );
  }
}
