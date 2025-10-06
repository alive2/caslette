import 'package:flutter/material.dart' hide Card;
import '../models/poker_models.dart';

class CardWidget extends StatelessWidget {
  final Card card;
  final double width;
  final double height;
  final bool faceDown;

  const CardWidget({
    super.key,
    required this.card,
    this.width = 60,
    this.height = 84,
    this.faceDown = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: width,
      height: height,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.black, width: 1),
        color: faceDown ? Colors.blue[900] : Colors.white,
        boxShadow: [
          BoxShadow(
            color: Colors.black.withOpacity(0.2),
            blurRadius: 4,
            offset: const Offset(2, 2),
          ),
        ],
      ),
      child: faceDown
          ? Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                color: Colors.blue[900],
              ),
              child: const Center(
                child: Icon(Icons.casino, color: Colors.white, size: 20),
              ),
            )
          : Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Text(
                    _getRankSymbol(card.rank),
                    style: TextStyle(
                      fontSize: width * 0.3,
                      fontWeight: FontWeight.bold,
                      color: _isRed(card.suit) ? Colors.red : Colors.black,
                    ),
                  ),
                  Text(
                    _getSuitSymbol(card.suit),
                    style: TextStyle(
                      fontSize: width * 0.25,
                      color: _isRed(card.suit) ? Colors.red : Colors.black,
                    ),
                  ),
                ],
              ),
            ),
    );
  }

  String _getRankSymbol(String rank) {
    switch (rank.toLowerCase()) {
      case 'jack':
        return 'J';
      case 'queen':
        return 'Q';
      case 'king':
        return 'K';
      case 'ace':
        return 'A';
      default:
        return rank;
    }
  }

  String _getSuitSymbol(String suit) {
    switch (suit.toLowerCase()) {
      case 'hearts':
        return '♥';
      case 'diamonds':
        return '♦';
      case 'clubs':
        return '♣';
      case 'spades':
        return '♠';
      default:
        return suit.substring(0, 1).toUpperCase();
    }
  }

  bool _isRed(String suit) {
    return suit.toLowerCase() == 'hearts' || suit.toLowerCase() == 'diamonds';
  }
}

class HandWidget extends StatelessWidget {
  final List<Card> cards;
  final double cardWidth;
  final double cardHeight;
  final bool faceDown;
  final double overlap;

  const HandWidget({
    super.key,
    required this.cards,
    this.cardWidth = 60,
    this.cardHeight = 84,
    this.faceDown = false,
    this.overlap = 20,
  });

  @override
  Widget build(BuildContext context) {
    if (cards.isEmpty) {
      return SizedBox(width: cardWidth, height: cardHeight);
    }

    return SizedBox(
      width: cardWidth + (cards.length - 1) * overlap,
      height: cardHeight,
      child: Stack(
        children: cards.asMap().entries.map((entry) {
          final index = entry.key;
          final card = entry.value;
          return Positioned(
            left: index * overlap,
            child: CardWidget(
              card: card,
              width: cardWidth,
              height: cardHeight,
              faceDown: faceDown,
            ),
          );
        }).toList(),
      ),
    );
  }
}

class CommunityCardsWidget extends StatelessWidget {
  final List<Card> cards;
  final double cardWidth;
  final double cardHeight;
  final double spacing;

  const CommunityCardsWidget({
    super.key,
    required this.cards,
    this.cardWidth = 70,
    this.cardHeight = 98,
    this.spacing = 8,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.green[800],
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.green[600]!, width: 2),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Show 5 card slots
          for (int i = 0; i < 5; i++) ...[
            if (i > 0) SizedBox(width: spacing),
            if (i < cards.length)
              CardWidget(card: cards[i], width: cardWidth, height: cardHeight)
            else
              Container(
                width: cardWidth,
                height: cardHeight,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.grey[400]!, width: 2),
                  color: Colors.grey[200],
                ),
                child: Icon(
                  Icons.crop_portrait,
                  color: Colors.grey[400],
                  size: cardWidth * 0.3,
                ),
              ),
          ],
        ],
      ),
    );
  }
}

class PotWidget extends StatelessWidget {
  final int pot;
  final int currentBet;

  const PotWidget({super.key, required this.pot, this.currentBet = 0});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.amber[100],
        borderRadius: BorderRadius.circular(20),
        border: Border.all(color: Colors.amber[800]!, width: 2),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.monetization_on, color: Colors.amber[800], size: 20),
              const SizedBox(width: 4),
              Text(
                'Pot: \$${pot.toString()}',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Colors.amber[800],
                ),
              ),
            ],
          ),
          if (currentBet > 0) ...[
            const SizedBox(height: 4),
            Text(
              'Current Bet: \$${currentBet.toString()}',
              style: TextStyle(fontSize: 14, color: Colors.amber[700]),
            ),
          ],
        ],
      ),
    );
  }
}

class PlayerWidget extends StatelessWidget {
  final PokerPlayer player;
  final bool isActive;
  final bool isDealer;
  final bool isSmallBlind;
  final bool isBigBlind;
  final bool isCurrentUser;

  const PlayerWidget({
    super.key,
    required this.player,
    this.isActive = false,
    this.isDealer = false,
    this.isSmallBlind = false,
    this.isBigBlind = false,
    this.isCurrentUser = false,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 120,
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: isCurrentUser
            ? Colors.blue[100]
            : isActive
            ? Colors.green[100]
            : Colors.grey[100],
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: isCurrentUser
              ? Colors.blue[600]!
              : isActive
              ? Colors.green[600]!
              : Colors.grey[400]!,
          width: isActive ? 3 : 1,
        ),
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // Player name and badges
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              if (isDealer) ...[
                Container(
                  padding: const EdgeInsets.all(2),
                  decoration: BoxDecoration(
                    color: Colors.orange[600],
                    shape: BoxShape.circle,
                  ),
                  child: const Text(
                    'D',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 10,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const SizedBox(width: 4),
              ],
              if (isSmallBlind) ...[
                Container(
                  padding: const EdgeInsets.all(2),
                  decoration: BoxDecoration(
                    color: Colors.blue[600],
                    shape: BoxShape.circle,
                  ),
                  child: const Text(
                    'SB',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 8,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const SizedBox(width: 4),
              ],
              if (isBigBlind) ...[
                Container(
                  padding: const EdgeInsets.all(2),
                  decoration: BoxDecoration(
                    color: Colors.red[600],
                    shape: BoxShape.circle,
                  ),
                  child: const Text(
                    'BB',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 8,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
                const SizedBox(width: 4),
              ],
            ],
          ),
          const SizedBox(height: 4),

          // Player name
          Text(
            player.username,
            style: TextStyle(
              fontWeight: FontWeight.bold,
              fontSize: 12,
              color: player.isFolded ? Colors.grey[600] : Colors.black,
            ),
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
          ),

          const SizedBox(height: 4),

          // Player cards (only show if current user or showdown)
          if (player.holeCards.isNotEmpty)
            HandWidget(
              cards: player.holeCards,
              cardWidth: 30,
              cardHeight: 42,
              faceDown: !isCurrentUser,
              overlap: 10,
            )
          else
            const SizedBox(height: 42),

          const SizedBox(height: 4),

          // Chips
          Text(
            '\$${player.chips}',
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.w600,
              color: player.chips > 0 ? Colors.green[700] : Colors.red[700],
            ),
          ),

          // Current bet
          if (player.currentBet > 0) ...[
            const SizedBox(height: 2),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: Colors.amber[200],
                borderRadius: BorderRadius.circular(10),
              ),
              child: Text(
                'Bet: \$${player.currentBet}',
                style: const TextStyle(
                  fontSize: 10,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ),
          ],

          // Last action
          if (player.lastAction.isNotEmpty) ...[
            const SizedBox(height: 2),
            Text(
              player.lastAction.toUpperCase(),
              style: TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: _getActionColor(player.lastAction),
              ),
            ),
          ],

          // Status indicators
          if (player.isFolded) ...[
            const SizedBox(height: 2),
            const Text(
              'FOLDED',
              style: TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: Colors.red,
              ),
            ),
          ] else if (player.isAllIn) ...[
            const SizedBox(height: 2),
            const Text(
              'ALL-IN',
              style: TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: Colors.purple,
              ),
            ),
          ] else if (!player.isReady && !isActive) ...[
            const SizedBox(height: 2),
            const Text(
              'NOT READY',
              style: TextStyle(
                fontSize: 10,
                fontWeight: FontWeight.bold,
                color: Colors.orange,
              ),
            ),
          ],
        ],
      ),
    );
  }

  Color _getActionColor(String action) {
    switch (action.toLowerCase()) {
      case 'fold':
        return Colors.red;
      case 'call':
        return Colors.blue;
      case 'raise':
      case 'bet':
        return Colors.green;
      case 'check':
        return Colors.orange;
      case 'all_in':
        return Colors.purple;
      default:
        return Colors.grey;
    }
  }
}
