enum Suit {
  hearts,
  diamonds,
  clubs,
  spades;

  String get symbol {
    switch (this) {
      case Suit.hearts:
        return '♥';
      case Suit.diamonds:
        return '♦';
      case Suit.clubs:
        return '♣';
      case Suit.spades:
        return '♠';
    }
  }

  String get name {
    switch (this) {
      case Suit.hearts:
        return 'Hearts';
      case Suit.diamonds:
        return 'Diamonds';
      case Suit.clubs:
        return 'Clubs';
      case Suit.spades:
        return 'Spades';
    }
  }

  bool get isRed => this == Suit.hearts || this == Suit.diamonds;
  bool get isBlack => this == Suit.clubs || this == Suit.spades;
}

enum Rank {
  two,
  three,
  four,
  five,
  six,
  seven,
  eight,
  nine,
  ten,
  jack,
  queen,
  king,
  ace;

  String get symbol {
    switch (this) {
      case Rank.two:
        return '2';
      case Rank.three:
        return '3';
      case Rank.four:
        return '4';
      case Rank.five:
        return '5';
      case Rank.six:
        return '6';
      case Rank.seven:
        return '7';
      case Rank.eight:
        return '8';
      case Rank.nine:
        return '9';
      case Rank.ten:
        return '10';
      case Rank.jack:
        return 'J';
      case Rank.queen:
        return 'Q';
      case Rank.king:
        return 'K';
      case Rank.ace:
        return 'A';
    }
  }

  String get name {
    switch (this) {
      case Rank.two:
        return 'Two';
      case Rank.three:
        return 'Three';
      case Rank.four:
        return 'Four';
      case Rank.five:
        return 'Five';
      case Rank.six:
        return 'Six';
      case Rank.seven:
        return 'Seven';
      case Rank.eight:
        return 'Eight';
      case Rank.nine:
        return 'Nine';
      case Rank.ten:
        return 'Ten';
      case Rank.jack:
        return 'Jack';
      case Rank.queen:
        return 'Queen';
      case Rank.king:
        return 'King';
      case Rank.ace:
        return 'Ace';
    }
  }

  int get value => index + 2; // 2-14 for poker calculations
}

class Card {
  final Suit suit;
  final Rank rank;

  const Card({required this.suit, required this.rank});

  factory Card.fromJson(Map<String, dynamic> json) {
    return Card(
      suit: Suit.values[json['suit']],
      rank: Rank.values[json['rank']],
    );
  }

  Map<String, dynamic> toJson() {
    return {'suit': suit.index, 'rank': rank.index};
  }

  factory Card.fromString(String cardString) {
    // Parse strings like "AH" (Ace of Hearts), "2C" (Two of Clubs)
    if (cardString.length < 2) {
      throw ArgumentError('Invalid card string: $cardString');
    }

    final rankChar = cardString[0];
    final suitChar = cardString[1];

    Rank rank;
    switch (rankChar.toUpperCase()) {
      case '2':
        rank = Rank.two;
        break;
      case '3':
        rank = Rank.three;
        break;
      case '4':
        rank = Rank.four;
        break;
      case '5':
        rank = Rank.five;
        break;
      case '6':
        rank = Rank.six;
        break;
      case '7':
        rank = Rank.seven;
        break;
      case '8':
        rank = Rank.eight;
        break;
      case '9':
        rank = Rank.nine;
        break;
      case 'T':
      case '10':
        rank = Rank.ten;
        break;
      case 'J':
        rank = Rank.jack;
        break;
      case 'Q':
        rank = Rank.queen;
        break;
      case 'K':
        rank = Rank.king;
        break;
      case 'A':
        rank = Rank.ace;
        break;
      default:
        throw ArgumentError('Invalid rank: $rankChar');
    }

    Suit suit;
    switch (suitChar.toUpperCase()) {
      case 'H':
        suit = Suit.hearts;
        break;
      case 'D':
        suit = Suit.diamonds;
        break;
      case 'C':
        suit = Suit.clubs;
        break;
      case 'S':
        suit = Suit.spades;
        break;
      default:
        throw ArgumentError('Invalid suit: $suitChar');
    }

    return Card(suit: suit, rank: rank);
  }

  @override
  String toString() {
    final rankStr = rank == Rank.ten ? 'T' : rank.symbol;
    final suitStr = suit.symbol[0].toUpperCase();
    return '$rankStr$suitStr';
  }

  String get displayString => '${rank.symbol}${suit.symbol}';

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is Card &&
          runtimeType == other.runtimeType &&
          suit == other.suit &&
          rank == other.rank;

  @override
  int get hashCode => suit.hashCode ^ rank.hashCode;
}
