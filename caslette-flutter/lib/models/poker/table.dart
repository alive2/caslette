import 'card.dart';
import 'player.dart';

enum BettingRound {
  preflop,
  flop,
  turn,
  river,
  showdown;

  String get displayName {
    switch (this) {
      case BettingRound.preflop:
        return 'Pre-flop';
      case BettingRound.flop:
        return 'Flop';
      case BettingRound.turn:
        return 'Turn';
      case BettingRound.river:
        return 'River';
      case BettingRound.showdown:
        return 'Showdown';
    }
  }
}

enum TableStatus {
  waiting,
  active,
  finished;

  String get displayName {
    switch (this) {
      case TableStatus.waiting:
        return 'Waiting for Players';
      case TableStatus.active:
        return 'In Progress';
      case TableStatus.finished:
        return 'Finished';
    }
  }
}

enum PlayerAction {
  fold,
  check,
  call,
  bet,
  raise,
  allin;

  String get displayName {
    switch (this) {
      case PlayerAction.fold:
        return 'Fold';
      case PlayerAction.check:
        return 'Check';
      case PlayerAction.call:
        return 'Call';
      case PlayerAction.bet:
        return 'Bet';
      case PlayerAction.raise:
        return 'Raise';
      case PlayerAction.allin:
        return 'All-in';
    }
  }
}

class PokerTable {
  final String id;
  final String name;
  final int maxPlayers;
  final int currentPlayers;
  final int smallBlind;
  final int bigBlind;
  final int minBuyIn;
  final int maxBuyIn;
  final TableStatus status;
  final List<PokerPlayer> players;
  final BettingRound bettingRound;
  final List<Card> communityCards;
  final int pot;
  final int currentBet;
  final String? currentPlayerUserId;
  final int handNumber;
  final Map<String, dynamic> metadata;

  const PokerTable({
    required this.id,
    required this.name,
    required this.maxPlayers,
    required this.currentPlayers,
    required this.smallBlind,
    required this.bigBlind,
    required this.minBuyIn,
    required this.maxBuyIn,
    required this.status,
    this.players = const [],
    this.bettingRound = BettingRound.preflop,
    this.communityCards = const [],
    this.pot = 0,
    this.currentBet = 0,
    this.currentPlayerUserId,
    this.handNumber = 0,
    this.metadata = const {},
  });

  factory PokerTable.fromJson(Map<String, dynamic> json) {
    return PokerTable(
      id: json['id']?.toString() ?? '',
      name: json['name']?.toString() ?? '',
      maxPlayers: json['max_players'] ?? 8,
      currentPlayers: json['current_players'] ?? 0,
      smallBlind: json['small_blind'] ?? 1,
      bigBlind: json['big_blind'] ?? 2,
      minBuyIn: json['min_buy_in'] ?? 100,
      maxBuyIn: json['max_buy_in'] ?? 1000,
      status: TableStatus.values.firstWhere(
        (status) => status.name == json['status'],
        orElse: () => TableStatus.waiting,
      ),
      players:
          (json['players'] as List<dynamic>?)
              ?.map((playerJson) => PokerPlayer.fromJson(playerJson))
              .toList() ??
          [],
      bettingRound: BettingRound.values.firstWhere(
        (round) => round.name == json['betting_round'],
        orElse: () => BettingRound.preflop,
      ),
      communityCards:
          (json['community_cards'] as List<dynamic>?)
              ?.map((cardJson) => Card.fromJson(cardJson))
              .toList() ??
          [],
      pot: json['pot'] ?? 0,
      currentBet: json['current_bet'] ?? 0,
      currentPlayerUserId: json['current_player_user_id']?.toString(),
      handNumber: json['hand_number'] ?? 0,
      metadata: json['metadata'] ?? {},
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'name': name,
      'max_players': maxPlayers,
      'current_players': currentPlayers,
      'small_blind': smallBlind,
      'big_blind': bigBlind,
      'min_buy_in': minBuyIn,
      'max_buy_in': maxBuyIn,
      'status': status.name,
      'players': players.map((player) => player.toJson()).toList(),
      'betting_round': bettingRound.name,
      'community_cards': communityCards.map((card) => card.toJson()).toList(),
      'pot': pot,
      'current_bet': currentBet,
      'current_player_user_id': currentPlayerUserId,
      'hand_number': handNumber,
      'metadata': metadata,
    };
  }

  PokerPlayer? getPlayerByUserId(String userId) {
    try {
      return players.firstWhere((player) => player.userId == userId);
    } catch (e) {
      return null;
    }
  }

  PokerPlayer? getCurrentPlayer() {
    if (currentPlayerUserId == null) return null;
    return getPlayerByUserId(currentPlayerUserId!);
  }

  bool get isWaitingForPlayers =>
      status == TableStatus.waiting || currentPlayers < 2;

  bool get isGameActive => status == TableStatus.active;

  bool get hasAvailableSeats => currentPlayers < maxPlayers;

  List<int> get availableSeats {
    final occupiedSeats = players.map((p) => p.seatNumber).toSet();
    return List.generate(
      maxPlayers,
      (index) => index + 1,
    ).where((seatNumber) => !occupiedSeats.contains(seatNumber)).toList();
  }

  PokerTable copyWith({
    String? id,
    String? name,
    int? maxPlayers,
    int? currentPlayers,
    int? smallBlind,
    int? bigBlind,
    int? minBuyIn,
    int? maxBuyIn,
    TableStatus? status,
    List<PokerPlayer>? players,
    BettingRound? bettingRound,
    List<Card>? communityCards,
    int? pot,
    int? currentBet,
    String? currentPlayerUserId,
    int? handNumber,
    Map<String, dynamic>? metadata,
  }) {
    return PokerTable(
      id: id ?? this.id,
      name: name ?? this.name,
      maxPlayers: maxPlayers ?? this.maxPlayers,
      currentPlayers: currentPlayers ?? this.currentPlayers,
      smallBlind: smallBlind ?? this.smallBlind,
      bigBlind: bigBlind ?? this.bigBlind,
      minBuyIn: minBuyIn ?? this.minBuyIn,
      maxBuyIn: maxBuyIn ?? this.maxBuyIn,
      status: status ?? this.status,
      players: players ?? this.players,
      bettingRound: bettingRound ?? this.bettingRound,
      communityCards: communityCards ?? this.communityCards,
      pot: pot ?? this.pot,
      currentBet: currentBet ?? this.currentBet,
      currentPlayerUserId: currentPlayerUserId ?? this.currentPlayerUserId,
      handNumber: handNumber ?? this.handNumber,
      metadata: metadata ?? this.metadata,
    );
  }

  @override
  bool operator ==(Object other) =>
      identical(this, other) ||
      other is PokerTable && runtimeType == other.runtimeType && id == other.id;

  @override
  int get hashCode => id.hashCode;
}
