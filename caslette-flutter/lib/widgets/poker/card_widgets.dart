import 'package:flutter/material.dart';
import '../../models/poker/poker_models.dart' as poker;

class PlayingCard extends StatelessWidget {
  final poker.Card? card;
  final bool isRevealed;
  final double width;
  final double height;
  final bool isSelected;
  final VoidCallback? onTap;

  const PlayingCard({
    super.key,
    this.card,
    this.isRevealed = true,
    this.width = 60,
    this.height = 84,
    this.isSelected = false,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: Container(
        width: width,
        height: height,
        decoration: BoxDecoration(
          color: isRevealed && card != null
              ? Colors.white
              : const Color(0xFF1E40AF),
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isSelected
                ? const Color(0xFF9333EA)
                : const Color(0xFF374151),
            width: isSelected ? 2 : 1,
          ),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.2),
              blurRadius: 4,
              offset: const Offset(0, 2),
            ),
          ],
        ),
        child: isRevealed && card != null
            ? _buildRevealedCard()
            : _buildCardBack(),
      ),
    );
  }

  Widget _buildRevealedCard() {
    if (card == null) return const SizedBox();

    final isRed = card!.suit.isRed;
    final color = isRed ? Colors.red : Colors.black;

    return Column(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        // Top left rank and suit
        Padding(
          padding: const EdgeInsets.all(4),
          child: Align(
            alignment: Alignment.topLeft,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  card!.rank.symbol,
                  style: TextStyle(
                    color: color,
                    fontSize: width * 0.2,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  card!.suit.symbol,
                  style: TextStyle(color: color, fontSize: width * 0.15),
                ),
              ],
            ),
          ),
        ),
        // Center suit symbol
        Text(
          card!.suit.symbol,
          style: TextStyle(color: color, fontSize: width * 0.4),
        ),
        // Bottom right rank and suit (rotated)
        Padding(
          padding: const EdgeInsets.all(4),
          child: Align(
            alignment: Alignment.bottomRight,
            child: Transform.rotate(
              angle: 3.14159, // 180 degrees
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    card!.rank.symbol,
                    style: TextStyle(
                      color: color,
                      fontSize: width * 0.2,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Text(
                    card!.suit.symbol,
                    style: TextStyle(color: color, fontSize: width * 0.15),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildCardBack() {
    return Container(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [Color(0xFF1E40AF), Color(0xFF3B82F6)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Center(
        child: Icon(
          Icons.casino,
          color: Colors.white.withOpacity(0.3),
          size: width * 0.5,
        ),
      ),
    );
  }
}

class CommunityCards extends StatelessWidget {
  final List<poker.Card> cards;
  final poker.BettingRound bettingRound;
  final double cardWidth;
  final double cardHeight;

  const CommunityCards({
    super.key,
    required this.cards,
    required this.bettingRound,
    this.cardWidth = 50,
    this.cardHeight = 70,
  });

  @override
  Widget build(BuildContext context) {
    // Determine how many cards should be visible based on betting round
    int visibleCards = 0;
    switch (bettingRound) {
      case poker.BettingRound.preflop:
        visibleCards = 0;
        break;
      case poker.BettingRound.flop:
        visibleCards = 3;
        break;
      case poker.BettingRound.turn:
        visibleCards = 4;
        break;
      case poker.BettingRound.river:
      case poker.BettingRound.showdown:
        visibleCards = 5;
        break;
    }

    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF059669).withOpacity(0.1),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: const Color(0xFF059669).withOpacity(0.3),
          width: 1,
        ),
      ),
      child: Column(
        children: [
          Text(
            'Community Cards',
            style: TextStyle(
              color: Colors.white.withOpacity(0.8),
              fontSize: 14,
              fontWeight: FontWeight.w500,
            ),
          ),
          const SizedBox(height: 12),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: List.generate(5, (index) {
              final shouldShow = index < visibleCards && index < cards.length;
              return Padding(
                padding: EdgeInsets.only(right: index < 4 ? 8 : 0),
                child: PlayingCard(
                  card: shouldShow ? cards[index] : null,
                  isRevealed: shouldShow,
                  width: cardWidth,
                  height: cardHeight,
                ),
              );
            }),
          ),
        ],
      ),
    );
  }
}

class PlayerHoleCards extends StatelessWidget {
  final List<poker.Card> cards;
  final bool isRevealed;
  final double cardWidth;
  final double cardHeight;

  const PlayerHoleCards({
    super.key,
    required this.cards,
    this.isRevealed = true,
    this.cardWidth = 45,
    this.cardHeight = 63,
  });

  @override
  Widget build(BuildContext context) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        PlayingCard(
          card: cards.isNotEmpty ? cards[0] : null,
          isRevealed: isRevealed,
          width: cardWidth,
          height: cardHeight,
        ),
        const SizedBox(width: 4),
        PlayingCard(
          card: cards.length > 1 ? cards[1] : null,
          isRevealed: isRevealed,
          width: cardWidth,
          height: cardHeight,
        ),
      ],
    );
  }
}
