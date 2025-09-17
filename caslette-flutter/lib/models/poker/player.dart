import 'card.dart';

enum PlayerStatus {
  waiting,
  active,
  folded,
  allin,
  sitting_out;

  String get displayName {
    switch (this) {
      case PlayerStatus.waiting:
        return 'Waiting';
      case PlayerStatus.active:
        return 'Active';
      case PlayerStatus.folded:
        return 'Folded';
      case PlayerStatus.allin:
        return 'All-in';
      case PlayerStatus.sitting_out:
        return 'Sitting Out';
    }
  }
}

class PokerPlayer {
  final String userId;
  final String username;
  final int seatNumber;
  final int chipCount;
  final PlayerStatus status;
  final List<Card> holeCards;
  final int currentBet;
  final bool isDealer;
  final bool isSmallBlind;
  final bool isBigBlind;
  final bool isCurrentPlayer;

  const PokerPlayer({
    required this.userId,
    required this.username,
    required this.seatNumber,
    required this.chipCount,
    required this.status,
    this.holeCards = const [],
    this.currentBet = 0,
    this.isDealer = false,
    this.isSmallBlind = false,
    this.isBigBlind = false,
    this.isCurrentPlayer = false,
  });

  factory PokerPlayer.fromJson(Map<String, dynamic> json) {
    return PokerPlayer(
      userId: json['user_id'] ?? '',
      username: json['username'] ?? '',
      seatNumber: json['seat_number'] ?? 0,
      chipCount: json['chip_count'] ?? 0,
      status: PlayerStatus.values.firstWhere(
        (status) => status.name == json['status'],
        orElse: () => PlayerStatus.waiting,
      ),
      holeCards:
          (json['hole_cards'] as List<dynamic>?)
              ?.map((cardJson) => Card.fromJson(cardJson))
              .toList() ??
          [],
      currentBet: json['current_bet'] ?? 0,
      isDealer: json['is_dealer'] ?? false,
      isSmallBlind: json['is_small_blind'] ?? false,
      isBigBlind: json['is_big_blind'] ?? false,
      isCurrentPlayer: json['is_current_player'] ?? false,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'user_id': userId,
      'username': username,
      'seat_number': seatNumber,
      'chip_count': chipCount,
      'status': status.name,
      'hole_cards': holeCards.map((card) => card.toJson()).toList(),
      'current_bet': currentBet,
      'is_dealer': isDealer,
      'is_small_blind': isSmallBlind,
      'is_big_blind': isBigBlind,
      'is_current_player': isCurrentPlayer,
    };
  }

  PokerPlayer copyWith({
    String? userId,
    String? username,
    int? seatNumber,
    int? chipCount,
    PlayerStatus? status,
    List<Card>? holeCards,
    int? currentBet,
    bool? isDealer,
    bool? isSmallBlind,
    bool? isBigBlind,
    bool? isCurrentPlayer,
  }) {
    return PokerPlayer(
      userId: userId ?? this.userId,
      username: username ?? this.username,
      seatNumber: seatNumber ?? this.seatNumber,
      chipCount: chipCount ?? this.chipCount,
      status: status ?? this.status,
      holeCards: holeCards ?? this.holeCards,
      currentBet: currentBet ?? this.currentBet,
      isDealer: isDealer ?? this.isDealer,
      isSmallBlind: isSmallBlind ?? this.isSmallBlind,
      isBigBlind: isBigBlind ?? this.isBigBlind,
      isCurrentPlayer: isCurrentPlayer ?? this.isCurrentPlayer,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is PokerPlayer &&
          runtimeType == other.runtimeType &&
          userId == other.userId &&
          seatNumber == other.seatNumber;

  @override
  int get hashCode => userId.hashCode ^ seatNumber.hashCode;
}
